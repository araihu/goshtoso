package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	captureServerChallengeHeader     = "X-Goshtoso-T-GS-011-AT-Challenge"
	captureServerCandidateTreeHeader = "X-Goshtoso-T-GS-011-Candidate-Tree"
	captureServerManifestHeader      = "X-Goshtoso-T-GS-011-Manifest-SHA256"
	captureServerBodySHA256Header    = "X-Goshtoso-T-GS-011-Body-SHA256"
	captureServerPairHeader          = "X-Goshtoso-T-GS-011-AT-Pair"
	captureServerQuery               = "t-gs-011-at-capture"
	captureServerStateQuery          = "t-gs-011-at-state"
	captureServerActionTokenQuery    = "t-gs-011-at-action-token"
	captureServedPageSchema          = "goshtoso.t-gs-011.at-served-page.v1"
)

// servedCandidatePage is derived from one direct response from the capture
// adapter's own loopback server. The raw HTML carries the challenge and source
// identity in addition to this metadata record.
type servedCandidatePage struct {
	Schema         string `json:"schema"`
	URL            string `json:"url"`
	Status         int    `json:"status"`
	Challenge      string `json:"challenge"`
	CandidateTree  string `json:"candidate_tree"`
	ManifestSHA256 string `json:"manifest_sha256"`
	BodySHA256     string `json:"body_sha256"`
	Pair           string `json:"pair"`
	ActionState    string `json:"action_state"`
	ActionToken    string `json:"action_token"`
}

type verifiedCandidateServer struct {
	URL             *url.URL
	command         *exec.Cmd
	workspace       string
	processOutput   bytes.Buffer
	requestTimeout  time.Duration
	shutdownTimeout time.Duration
	pair            string
}

func startVerifiedCandidateServer(repositoryRoot string, identity candidateIdentity, challenge captureChallenge, pair string) (*verifiedCandidateServer, error) {
	if err := verifyCandidateIdentity(repositoryRoot, identity); err != nil {
		return nil, fmt.Errorf("verify frozen candidate before server launch: %w", err)
	}
	workspace, err := os.MkdirTemp("", "goshtoso-scrollregion-at-candidate-server-")
	if err != nil {
		return nil, fmt.Errorf("create isolated candidate server workspace: %w", err)
	}
	cleanup := func(cause error) (*verifiedCandidateServer, error) {
		_ = os.RemoveAll(workspace)
		return nil, cause
	}
	workfile := filepath.Join(workspace, "go.work")
	workfileBody := fmt.Appendf(nil, "go %s\n\nuse (\n\t%s\n\t%s\n)\n", identity.DependencyPins.RootGoDirective, repositoryRoot, filepath.Join(repositoryRoot, "site"))
	if err := os.WriteFile(workfile, workfileBody, 0o600); err != nil {
		return cleanup(fmt.Errorf("write isolated candidate server workspace: %w", err))
	}
	binaryPath := filepath.Join(workspace, "scrollregion-candidate-server")
	build := exec.Command("go", "build", "-o", binaryPath, "./site/cmd/server")
	build.Dir = repositoryRoot
	build.Env = append(os.Environ(), "GOWORK="+workfile)
	if output, err := build.CombinedOutput(); err != nil {
		return cleanup(fmt.Errorf("build verified candidate server: %w: %s", err, strings.TrimSpace(string(output))))
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return cleanup(fmt.Errorf("reserve loopback candidate server port: %w", err))
	}
	port := listener.Addr().(*net.TCPAddr).Port
	if err := listener.Close(); err != nil {
		return cleanup(fmt.Errorf("release loopback candidate server port: %w", err))
	}
	serverURL, err := url.Parse(fmt.Sprintf("http://127.0.0.1:%d%s?%s=%s", port, captureRoute, captureServerQuery, challenge.Challenge))
	if err != nil {
		return cleanup(fmt.Errorf("build loopback candidate URL: %w", err))
	}
	server := &verifiedCandidateServer{URL: serverURL, workspace: workspace, requestTimeout: 12 * time.Second, shutdownTimeout: 5 * time.Second, pair: pair}
	server.URL = server.actionURL("default", challenge)
	command := exec.Command(binaryPath, "-port", strconv.Itoa(port))
	command.Dir = repositoryRoot
	command.Env = append(os.Environ(),
		"GOWORK="+workfile,
		"GOSHTOSO_PROJECT_ROOT="+repositoryRoot,
		"GOSHTOSO_SCROLLREGION_AT_CAPTURE_CHALLENGE="+challenge.Challenge,
		"GOSHTOSO_SCROLLREGION_AT_CAPTURE_CANDIDATE_TREE="+identity.CandidateTree,
		"GOSHTOSO_SCROLLREGION_AT_CAPTURE_MANIFEST_SHA256="+identity.ManifestSHA256,
		"GOSHTOSO_SCROLLREGION_AT_CAPTURE_PAIR="+pair,
	)
	command.Stdout = &server.processOutput
	command.Stderr = &server.processOutput
	if err := command.Start(); err != nil {
		return cleanup(fmt.Errorf("start verified candidate server: %w", err))
	}
	server.command = command
	return server, nil
}

// actionURL is adapter-owned. Every fixed AT state receives a separate
// challenge-derived token rendered by the verified candidate server; callers
// cannot choose an arbitrary origin, state, or token.
func (server *verifiedCandidateServer) actionURL(state string, challenge captureChallenge) *url.URL {
	copyURL := *server.URL
	query := copyURL.Query()
	query.Set(captureServerQuery, challenge.Challenge)
	query.Set(captureServerStateQuery, state)
	query.Set(captureServerActionTokenQuery, captureActionToken(challenge.Challenge, state))
	copyURL.RawQuery = query.Encode()
	return &copyURL
}

func captureActionToken(challenge, state string) string {
	return "goshtoso-t-gs-011-" + challenge + "-" + state
}

func (server *verifiedCandidateServer) fetch(identity candidateIdentity, challenge captureChallenge) (servedCandidatePage, []byte, error) {
	if server == nil || server.URL == nil || server.command == nil || server.command.Process == nil {
		return servedCandidatePage{}, nil, fmt.Errorf("verified candidate server is not running")
	}
	deadline := time.Now().Add(server.requestTimeout)
	client := &http.Client{Timeout: 2 * time.Second}
	var lastErr error
	for time.Now().Before(deadline) {
		request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL.String(), nil)
		if err != nil {
			return servedCandidatePage{}, nil, fmt.Errorf("create direct candidate request: %w", err)
		}
		response, err := client.Do(request)
		if err == nil {
			body, readErr := io.ReadAll(response.Body)
			_ = response.Body.Close()
			if readErr != nil {
				return servedCandidatePage{}, nil, fmt.Errorf("read direct candidate response: %w", readErr)
			}
			page := servedCandidatePage{
				Schema: captureServedPageSchema, URL: server.URL.String(), Status: response.StatusCode,
				Challenge: response.Header.Get(captureServerChallengeHeader), CandidateTree: response.Header.Get(captureServerCandidateTreeHeader),
				ManifestSHA256: response.Header.Get(captureServerManifestHeader), BodySHA256: response.Header.Get(captureServerBodySHA256Header), Pair: response.Header.Get(captureServerPairHeader),
				ActionState: "default", ActionToken: captureActionToken(challenge.Challenge, "default"),
			}
			if err := page.validate(identity, challenge, body); err != nil {
				return servedCandidatePage{}, nil, err
			}
			if page.Pair != server.pair {
				return servedCandidatePage{}, nil, fmt.Errorf("direct candidate response pair mismatch")
			}
			return page, body, nil
		}
		lastErr = err
		if server.command.ProcessState != nil && server.command.ProcessState.Exited() {
			return servedCandidatePage{}, nil, fmt.Errorf("verified candidate server exited before response: %s", strings.TrimSpace(server.processOutput.String()))
		}
		time.Sleep(150 * time.Millisecond)
	}
	return servedCandidatePage{}, nil, fmt.Errorf("wait for verified candidate server: %w: %s", lastErr, strings.TrimSpace(server.processOutput.String()))
}

func (page servedCandidatePage) validate(identity candidateIdentity, challenge captureChallenge, body []byte) error {
	if page.Schema != captureServedPageSchema || page.Status != http.StatusOK || page.Challenge != challenge.Challenge || page.CandidateTree != identity.CandidateTree || page.ManifestSHA256 != identity.ManifestSHA256 || page.Pair == "" || page.ActionState != "default" || page.ActionToken != captureActionToken(challenge.Challenge, page.ActionState) || !captureSHA256Pattern.MatchString(page.BodySHA256) {
		return fmt.Errorf("direct candidate response does not bind exact challenge and frozen source identity")
	}
	if page.BodySHA256 != sha256Hex(body) {
		return fmt.Errorf("direct candidate response body digest mismatch")
	}
	for _, required := range []string{
		`name="goshtoso-t-gs-011-at-challenge" content="` + challenge.Challenge + `"`,
		`name="goshtoso-t-gs-011-candidate-tree" content="` + identity.CandidateTree + `"`,
		`name="goshtoso-t-gs-011-manifest-sha256" content="` + identity.ManifestSHA256 + `"`,
		`name="goshtoso-t-gs-011-at-pair" content="` + page.Pair + `"`,
		`name="goshtoso-t-gs-011-at-action-state" content="` + page.ActionState + `"`,
		`name="goshtoso-t-gs-011-at-action-token" content="` + page.ActionToken + `"`,
		`data-goshtoso-scroll-viewport`,
		`role="region"`,
	} {
		if !bytes.Contains(body, []byte(required)) {
			return fmt.Errorf("direct candidate HTML lacks required bound content %q", required)
		}
	}
	return nil
}

func (server *verifiedCandidateServer) close() error {
	if server == nil {
		return nil
	}
	var result error
	if server.command != nil && server.command.Process != nil && (server.command.ProcessState == nil || !server.command.ProcessState.Exited()) {
		_ = server.command.Process.Signal(os.Interrupt)
		wait := make(chan error, 1)
		go func() { wait <- server.command.Wait() }()
		select {
		case err := <-wait:
			if err != nil {
				result = fmt.Errorf("stop verified candidate server: %w", err)
			}
		case <-time.After(server.shutdownTimeout):
			if err := server.command.Process.Kill(); err != nil {
				result = fmt.Errorf("stop timed-out verified candidate server: %w", err)
			}
			<-wait
		}
	}
	if server.workspace != "" {
		if err := os.RemoveAll(server.workspace); err != nil && result == nil {
			result = fmt.Errorf("remove isolated candidate server workspace: %w", err)
		}
	}
	return result
}
