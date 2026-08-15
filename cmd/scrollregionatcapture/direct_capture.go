package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const captureBrowserStateSchema = "goshtoso.t-gs-011.at-browser-state.v1"

const browserCaptureConfigMarker = "__GOSHTOSO_TGS011_CAPTURE_CONFIG__"

// The direct browser programs are authored JavaScript sources rather than Go
// string builders. Go supplies only the strictly encoded candidate binding and
// fixed action configuration immediately before the browser executes them.
//
//go:embed browser_prepare.js
var browserPrepareSource string

//go:embed browser_read_state.js
var browserReadStateSource string

//go:embed browser_window.js
var browserWindowSource string

type browserRectangle struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type browserStateCapture struct {
	Schema          string           `json:"schema"`
	Pair            string           `json:"pair"`
	State           string           `json:"state"`
	Command         string           `json:"command"`
	Route           string           `json:"route"`
	Challenge       string           `json:"challenge"`
	CandidateTree   string           `json:"candidate_tree"`
	ManifestSHA256  string           `json:"manifest_sha256"`
	ActionToken     string           `json:"action_token"`
	Phase           string           `json:"phase"`
	ObservedAt      string           `json:"observed_at"`
	Snapshot        browserSnapshot  `json:"snapshot"`
	CandidateRegion browserRectangle `json:"candidate_region"`
}

// browserSnapshot is emitted directly by the browser after each fixed action
// phase. It records actual active-element and scroll state; no DOM-neighbour
// inference is accepted as focus-navigation evidence.
type browserSnapshot struct {
	ActiveRole      string  `json:"active_role"`
	ActiveName      string  `json:"active_name"`
	RegionRole      string  `json:"region_role"`
	RegionName      string  `json:"region_name"`
	RegionFocused   bool    `json:"region_focused"`
	Boundary        string  `json:"boundary"`
	ScrollTop       float64 `json:"scroll_top"`
	ClientHeight    float64 `json:"client_height"`
	ScrollHeight    float64 `json:"scroll_height"`
	StartCueVisible bool    `json:"start_cue_visible"`
	EndCueVisible   bool    `json:"end_cue_visible"`
}

// captureActionRecord is generated only by the production adapter around an
// action it directly performs. The browser snapshots and scoped VoiceOver raw
// artifact independently echo its token and outcome.
type captureActionRecord struct {
	Schema             string `json:"schema"`
	Pair               string `json:"pair"`
	State              string `json:"state"`
	Command            string `json:"command"`
	Route              string `json:"route"`
	Challenge          string `json:"challenge"`
	CandidateTree      string `json:"candidate_tree"`
	ManifestSHA256     string `json:"manifest_sha256"`
	ActionToken        string `json:"action_token"`
	ActionEvent        string `json:"action_event"`
	ExitCommand        string `json:"exit_command"`
	ExitEvent          string `json:"exit_event"`
	VoiceOverPID       int    `json:"voiceover_pid"`
	VoiceOverSubsystem string `json:"voiceover_subsystem"`
	BeforeAt           string `json:"before_at"`
	LogStartedAt       string `json:"log_started_at"`
	ActionIssuedAt     string `json:"action_issued_at"`
	AfterAt            string `json:"after_at"`
	ExitIssuedAt       string `json:"exit_issued_at"`
	ExitAt             string `json:"exit_at"`
	LogEndedAt         string `json:"log_ended_at"`
}

const captureActionRecordSchema = "goshtoso.t-gs-011.at-action-record.v2"
const captureVoiceOverSubsystem = "com.apple.VoiceOver"
const captureExitCommand = "Tab"
const captureExitEvent = "macOS System Events key code 48"

// voiceOverLogEvent is the strict subset emitted by `log show --style json`
// that binds one VoiceOver utterance to one non-overlapping action interval.
// Plain prose/cumulative output is intentionally rejected.
type voiceOverLogEvent struct {
	Timestamp    string `json:"timestamp"`
	ProcessID    int    `json:"processID"`
	Subsystem    string `json:"subsystem"`
	EventMessage string `json:"eventMessage"`
}

type browserWindowCapture struct {
	Schema          string           `json:"schema"`
	Pair            string           `json:"pair"`
	Route           string           `json:"route"`
	Challenge       string           `json:"challenge"`
	CandidateTree   string           `json:"candidate_tree"`
	ManifestSHA256  string           `json:"manifest_sha256"`
	Window          browserRectangle `json:"window"`
	CandidateRegion browserRectangle `json:"candidate_region"`
}

func parseBrowserWindowCapture(raw []byte, plan capturePlan, identity candidateIdentity, challenge captureChallenge) (browserWindowCapture, error) {
	var capture browserWindowCapture
	if err := decodeStrictJSON(raw, &capture); err != nil {
		return capture, fmt.Errorf("decode direct browser window scope: %w", err)
	}
	if capture.Schema != captureBrowserStateSchema || capture.Pair != plan.Pair || capture.Route != captureRoute || capture.Challenge != challenge.Challenge || capture.CandidateTree != identity.CandidateTree || capture.ManifestSHA256 != identity.ManifestSHA256 ||
		capture.Window.Width < 100 || capture.Window.Height < 80 || capture.CandidateRegion.Width <= 0 || capture.CandidateRegion.Height <= 0 {
		return capture, fmt.Errorf("direct browser window scope does not bind exact candidate page and named region")
	}
	return capture, nil
}

func (capture browserWindowCapture) captureRect() string {
	x := int(math.Round(capture.Window.X))
	y := int(math.Round(capture.Window.Y))
	width := int(math.Round(capture.Window.Width))
	height := int(math.Round(capture.Window.Height))
	return fmt.Sprintf("%d,%d,%d,%d", x, y, width, height)
}

func captureFixedOutcomes(plan capturePlan, runCommand func(string, ...string) ([]byte, error), directory string, server *verifiedCandidateServer, identity candidateIdentity, challenge captureChallenge) ([]observation, []traceEvent, []evidenceArtifact, []evidenceArtifact, []evidenceArtifact, []byte, error) {
	if server == nil {
		return nil, nil, nil, nil, nil, nil, fmt.Errorf("direct capture requires an adapter-owned verified candidate server")
	}
	voiceOverPID, err := captureVoiceOverProcessID(runCommand)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}
	observations := make([]observation, 0, len(captureStates))
	events := make([]traceEvent, 0, len(captureStates))
	browserArtifacts := make([]evidenceArtifact, 0, len(captureStates)*3)
	voiceOverArtifacts := make([]evidenceArtifact, 0, len(captureStates))
	actionArtifacts := make([]evidenceArtifact, 0, len(captureStates))
	allVoiceOverEvents := make([]voiceOverLogEvent, 0, len(captureStates))
	var previousLogEnd time.Time
	for _, state := range captureStates {
		// The action token binds this independently challenged state to served
		// metadata, browser raw bytes, and the signed action record. It is never
		// announced into the tested DOM: VoiceOver must report the real region.
		token := captureActionToken(challenge.Challenge, state.Name)
		if _, err := runCommand("osascript", "-e", browserNavigateScript(plan, server.actionURL(state.Name, challenge).String())); err != nil {
			return nil, nil, nil, nil, nil, nil, fmt.Errorf("navigate direct browser to state %q: %w", state.Name, err)
		}
		time.Sleep(250 * time.Millisecond)
		if _, err := runCommand("osascript", "-e", browserPrepareScript(plan, state, identity, challenge)); err != nil {
			return nil, nil, nil, nil, nil, nil, fmt.Errorf("prepare direct browser state %q: %w", state.Name, err)
		}
		before, beforeArtifact, err := captureBrowserState(plan, state, "before", runCommand, directory, identity, challenge)
		if err != nil {
			return nil, nil, nil, nil, nil, nil, err
		}
		beforeAt, err := captureRFC3339Timestamp(before.ObservedAt, "before")
		if err != nil {
			return nil, nil, nil, nil, nil, nil, err
		}
		logStartedAt := captureActionNowAfter(beforeAt, previousLogEnd)
		actionIssuedAt := captureActionNowAfter(logStartedAt)
		if state.Name == "default" {
			// The default outcome is causally bound to a second real candidate
			// navigation after the before snapshot, never a page-load utterance
			// carried into a later artificial interval.
			if _, err := runCommand("osascript", "-e", browserNavigateScript(plan, server.actionURL(state.Name, challenge).String())); err != nil {
				return nil, nil, nil, nil, nil, nil, fmt.Errorf("perform direct browser navigation for default state: %w", err)
			}
			time.Sleep(250 * time.Millisecond)
		} else if state.KeyCode != 0 {
			if _, err := runCommand("osascript", "-e", browserKeyboardScript(plan, state.KeyCode)); err != nil {
				return nil, nil, nil, nil, nil, nil, fmt.Errorf("perform direct browser command %q for state %q: %w", state.Command, state.Name, err)
			}
		}
		// Let the system accessibility stream emit after the genuine keyboard
		// command. This is bounded per state; capture fails rather than accepting
		// claimant-entered text if the OS output never becomes observable.
		time.Sleep(250 * time.Millisecond)
		after, afterArtifact, err := captureBrowserState(plan, state, "after", runCommand, directory, identity, challenge)
		if err != nil {
			return nil, nil, nil, nil, nil, nil, err
		}
		afterAt, err := captureRFC3339Timestamp(after.ObservedAt, "after")
		if err != nil || !afterAt.After(actionIssuedAt) {
			return nil, nil, nil, nil, nil, nil, fmt.Errorf("direct browser after snapshot for state %q must occur after issued action: %w", state.Name, err)
		}
		exitIssuedAt := captureActionNowAfter(afterAt)
		if _, err := runCommand("osascript", "-e", browserKeyboardScript(plan, 48)); err != nil {
			return nil, nil, nil, nil, nil, nil, fmt.Errorf("perform direct browser focus-exit command for state %q: %w", state.Name, err)
		}
		time.Sleep(150 * time.Millisecond)
		exit, exitArtifact, err := captureBrowserState(plan, state, "exit", runCommand, directory, identity, challenge)
		if err != nil {
			return nil, nil, nil, nil, nil, nil, err
		}
		exitAt, err := captureRFC3339Timestamp(exit.ObservedAt, "exit")
		if err != nil || !exitAt.After(exitIssuedAt) {
			return nil, nil, nil, nil, nil, nil, fmt.Errorf("direct browser exit snapshot for state %q must occur after issued exit: %w", state.Name, err)
		}
		logEndedAt := captureActionNowAfter(exitAt)
		action := captureActionRecord{
			Schema: captureActionRecordSchema, Pair: plan.Pair, State: state.Name, Command: state.Command, Route: captureRoute,
			Challenge: challenge.Challenge, CandidateTree: identity.CandidateTree, ManifestSHA256: identity.ManifestSHA256,
			ActionToken: token, VoiceOverPID: voiceOverPID, VoiceOverSubsystem: captureVoiceOverSubsystem,
			ActionEvent: state.ActionEvent, ExitCommand: captureExitCommand, ExitEvent: captureExitEvent,
			BeforeAt: beforeAt.Format(time.RFC3339Nano), LogStartedAt: logStartedAt.Format(time.RFC3339Nano), ActionIssuedAt: actionIssuedAt.Format(time.RFC3339Nano),
			AfterAt: afterAt.Format(time.RFC3339Nano), ExitIssuedAt: exitIssuedAt.Format(time.RFC3339Nano), ExitAt: exitAt.Format(time.RFC3339Nano), LogEndedAt: logEndedAt.Format(time.RFC3339Nano),
		}
		actionPath := fmt.Sprintf("action-%s.json", state.Name)
		if err := writeJSONExclusive(filepathJoin(directory, actionPath), action); err != nil {
			return nil, nil, nil, nil, nil, nil, fmt.Errorf("write direct action record %q: %w", state.Name, err)
		}
		actionArtifact, err := artifactForFile(filepathJoin(directory, actionPath))
		if err != nil {
			return nil, nil, nil, nil, nil, nil, err
		}
		rawVoiceOver, err := runCommand("log", "show", "--style", "json", "--start", action.LogStartedAt, "--end", action.LogEndedAt, "--predicate", fmt.Sprintf(`processID == %d AND subsystem == %q`, action.VoiceOverPID, action.VoiceOverSubsystem))
		if err != nil {
			return nil, nil, nil, nil, nil, nil, fmt.Errorf("capture action-bound direct VoiceOver output for state %q: %w", state.Name, err)
		}
		voiceOverPath := fmt.Sprintf("voiceover-caption-%s.json", state.Name)
		if err := writeBytesExclusive(filepathJoin(directory, voiceOverPath), rawVoiceOver); err != nil {
			return nil, nil, nil, nil, nil, nil, fmt.Errorf("write direct VoiceOver output for state %q: %w", state.Name, err)
		}
		voiceOverArtifact, err := artifactForFile(filepathJoin(directory, voiceOverPath))
		if err != nil {
			return nil, nil, nil, nil, nil, nil, err
		}
		observed, err := deriveObservationFromRaw(state, before, after, exit, action, rawVoiceOver)
		if err != nil {
			return nil, nil, nil, nil, nil, nil, err
		}
		observations = append(observations, observed)
		events = append(events, traceEvent{State: state.Name, Command: state.Command, BeforeBoundary: before.Snapshot.Boundary, AfterBoundary: after.Snapshot.Boundary, Focused: after.Snapshot.RegionFocused})
		browserArtifacts = append(browserArtifacts, beforeArtifact, afterArtifact, exitArtifact)
		voiceOverArtifacts = append(voiceOverArtifacts, voiceOverArtifact)
		actionArtifacts = append(actionArtifacts, actionArtifact)
		voiceEvents, parseErr := parseVoiceOverLogEvents(rawVoiceOver)
		if parseErr != nil {
			return nil, nil, nil, nil, nil, nil, fmt.Errorf("decode action-bound direct VoiceOver output for state %q: %w", state.Name, parseErr)
		}
		allVoiceOverEvents = append(allVoiceOverEvents, voiceEvents...)
		previousLogEnd = logEndedAt
	}
	combinedVoiceOver, err := json.Marshal(allVoiceOverEvents)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, fmt.Errorf("encode combined action-bound VoiceOver output: %w", err)
	}
	return observations, events, browserArtifacts, voiceOverArtifacts, actionArtifacts, combinedVoiceOver, nil
}

func captureVoiceOverProcessID(runCommand func(string, ...string) ([]byte, error)) (int, error) {
	raw, err := runCommand("pgrep", "-x", "VoiceOver")
	if err != nil {
		return 0, fmt.Errorf("VoiceOver must be running before action-bound capture: %w", err)
	}
	lines := strings.Fields(string(raw))
	if len(lines) != 1 {
		return 0, fmt.Errorf("action-bound capture requires exactly one VoiceOver PID, got %q", strings.TrimSpace(string(raw)))
	}
	pid, err := strconv.Atoi(lines[0])
	if err != nil || pid < 2 {
		return 0, fmt.Errorf("action-bound capture received invalid VoiceOver PID %q", lines[0])
	}
	return pid, nil
}

func captureRFC3339Timestamp(value, phase string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil || parsed.Location() != time.UTC {
		return time.Time{}, fmt.Errorf("direct browser %s timestamp must be UTC RFC3339Nano", phase)
	}
	return parsed, nil
}

func captureActionNowAfter(previous ...time.Time) time.Time {
	now := time.Now().UTC()
	for _, value := range previous {
		if !value.IsZero() && !now.After(value) {
			now = value.Add(time.Nanosecond)
		}
	}
	return now
}

func captureBrowserState(plan capturePlan, state captureState, phase string, runCommand func(string, ...string) ([]byte, error), directory string, identity candidateIdentity, challenge captureChallenge) (browserStateCapture, evidenceArtifact, error) {
	raw, err := runCommand("osascript", "-e", browserReadStateScript(plan, state, phase, identity, challenge))
	if err != nil {
		return browserStateCapture{}, evidenceArtifact{}, fmt.Errorf("read direct browser state %q phase %q: %w", state.Name, phase, err)
	}
	browser, err := parseBrowserStateCapture(raw, state, phase, plan, identity, challenge)
	if err != nil {
		return browserStateCapture{}, evidenceArtifact{}, err
	}
	path := fmt.Sprintf("browser-state-%s-%s.json", state.Name, phase)
	if err := writeBytesExclusive(filepathJoin(directory, path), raw); err != nil {
		return browserStateCapture{}, evidenceArtifact{}, fmt.Errorf("write direct browser state %q phase %q: %w", state.Name, phase, err)
	}
	artifact, err := artifactForFile(filepathJoin(directory, path))
	if err != nil {
		return browserStateCapture{}, evidenceArtifact{}, err
	}
	return browser, artifact, nil
}

// filepathJoin is a narrow seam to keep direct evidence paths visibly rooted
// in the capture output directory at every call site.
func filepathJoin(directory, name string) string { return filepath.Join(directory, name) }

func parseBrowserStateCapture(raw []byte, state captureState, phase string, plan capturePlan, identity candidateIdentity, challenge captureChallenge) (browserStateCapture, error) {
	var capture browserStateCapture
	if err := decodeStrictJSON(raw, &capture); err != nil {
		return capture, fmt.Errorf("decode direct browser state %q: %w", state.Name, err)
	}
	if capture.Schema != captureBrowserStateSchema || capture.Pair != plan.Pair || capture.State != state.Name || capture.Command != state.Command || capture.Phase != phase || capture.ActionToken != captureActionToken(challenge.Challenge, state.Name) ||
		capture.Route != captureRoute || capture.Challenge != challenge.Challenge || capture.CandidateTree != identity.CandidateTree || capture.ManifestSHA256 != identity.ManifestSHA256 {
		return capture, fmt.Errorf("direct browser state %q does not bind exact command, challenge, and frozen candidate", state.Name)
	}
	if _, err := time.Parse(time.RFC3339Nano, capture.ObservedAt); err != nil || capture.Snapshot.RegionRole != "region" || !meaningfulCaptureText(capture.Snapshot.RegionName) ||
		!meaningfulCaptureText(capture.Snapshot.ActiveRole) || !meaningfulCaptureText(capture.Snapshot.ActiveName) || capture.Snapshot.ScrollTop < 0 || capture.Snapshot.ClientHeight <= 0 || capture.Snapshot.ScrollHeight < capture.Snapshot.ClientHeight ||
		capture.CandidateRegion.Width <= 0 || capture.CandidateRegion.Height <= 0 || math.IsNaN(capture.CandidateRegion.X) || math.IsNaN(capture.CandidateRegion.Y) {
		return capture, fmt.Errorf("direct browser state %q phase %q has incomplete observed active-element, scroll, role, or region geometry", state.Name, phase)
	}
	wantName := "Activity history"
	if state.Name == "no-overflow" {
		wantName = "Current release"
	}
	if capture.Snapshot.RegionName != wantName {
		return capture, fmt.Errorf("direct browser state %q named %q, want %q", state.Name, capture.Snapshot.RegionName, wantName)
	}
	return capture, nil
}

func deriveObservationFromRaw(state captureState, before, after, exit browserStateCapture, action captureActionRecord, voiceOver []byte) (observation, error) {
	if after.Snapshot.RegionFocused != state.Focused || after.Snapshot.Boundary != state.Boundary || before.ActionToken != action.ActionToken || after.ActionToken != action.ActionToken || exit.ActionToken != action.ActionToken {
		return observation{}, fmt.Errorf("direct browser snapshots do not prove state %q action outcome", state.Name)
	}
	if state.Focused && (after.Snapshot.ActiveRole != "region" || after.Snapshot.ActiveName != after.Snapshot.RegionName || exit.Snapshot.ActiveRole == "region") {
		return observation{}, fmt.Errorf("direct browser snapshots do not prove actual focus entry and exit for state %q", state.Name)
	}
	if err := validateCaptureActionTimeline(state, before, after, exit, action); err != nil {
		return observation{}, fmt.Errorf("direct browser snapshots do not prove causal action timing for state %q: %w", state.Name, err)
	}
	if err := validateCaptureStateTransition(state, before.Snapshot, after.Snapshot, exit.Snapshot); err != nil {
		return observation{}, fmt.Errorf("direct browser snapshots do not prove causal transition for state %q: %w", state.Name, err)
	}
	phrase, err := voiceOverPhrase(voiceOver, after.Snapshot.RegionName, action)
	if err != nil {
		return observation{}, fmt.Errorf("derive state %q from direct VoiceOver output: %w", state.Name, err)
	}
	navigation := focusNavigation{
		Before: snapshotDescription(before.Snapshot),
		Entry:  snapshotDescription(after.Snapshot),
		Exit:   snapshotDescription(exit.Snapshot),
	}
	return observation{
		State: state.Name, Role: after.Snapshot.RegionRole, Name: after.Snapshot.RegionName, Focused: after.Snapshot.RegionFocused, Boundary: after.Snapshot.Boundary,
		Commands: []string{state.Command}, FocusNavigation: navigation, ObservedSpeech: []string{phrase}, UnexpectedSpeech: []string{},
	}, nil
}

func validateCaptureActionTimeline(state captureState, before, after, exit browserStateCapture, action captureActionRecord) error {
	beforeObserved, err := captureRFC3339Timestamp(before.ObservedAt, "before observed")
	if err != nil {
		return err
	}
	afterObserved, err := captureRFC3339Timestamp(after.ObservedAt, "after observed")
	if err != nil {
		return err
	}
	exitObserved, err := captureRFC3339Timestamp(exit.ObservedAt, "exit observed")
	if err != nil {
		return err
	}
	values := []struct {
		name  string
		value string
	}{
		{"before_at", action.BeforeAt}, {"log_started_at", action.LogStartedAt}, {"action_issued_at", action.ActionIssuedAt},
		{"after_at", action.AfterAt}, {"exit_issued_at", action.ExitIssuedAt}, {"exit_at", action.ExitAt}, {"log_ended_at", action.LogEndedAt},
	}
	parsed := make([]time.Time, len(values))
	for index, entry := range values {
		parsed[index], err = captureRFC3339Timestamp(entry.value, entry.name)
		if err != nil {
			return err
		}
		if index > 0 && !parsed[index].After(parsed[index-1]) {
			return fmt.Errorf("timestamps are not strictly monotonic at %s", entry.name)
		}
	}
	if !beforeObserved.Equal(parsed[0]) || !afterObserved.Equal(parsed[3]) || !exitObserved.Equal(parsed[5]) {
		return fmt.Errorf("claimant action timestamps do not exactly match raw browser snapshot bytes")
	}
	if action.Command != state.Command || action.ActionEvent != state.ActionEvent || action.ExitCommand != captureExitCommand || action.ExitEvent != captureExitEvent {
		return fmt.Errorf("action contract does not match adapter-owned state command/event or Tab exit")
	}
	return nil
}

func validateCaptureStateTransition(state captureState, before, after, exit browserSnapshot) error {
	switch state.Name {
	case "start":
		if before.Boundary == "start" || after.Boundary != "start" || !after.RegionFocused {
			return fmt.Errorf("Home requires a non-start before boundary and focused start after boundary")
		}
	case "middle":
		if before.Boundary != "start" || after.Boundary != "middle" || !after.RegionFocused {
			return fmt.Errorf("PageDown requires a start before boundary and focused middle after boundary")
		}
	case "end":
		if before.Boundary == "end" || after.Boundary != "end" || !after.RegionFocused {
			return fmt.Errorf("End requires a non-end before boundary and focused end after boundary")
		}
	case "focused":
		if before.RegionFocused || before.ActiveRole == "region" || !after.RegionFocused || after.ActiveRole != "region" || exit.RegionFocused || exit.ActiveRole == "region" {
			return fmt.Errorf("Tab requires real outside-to-region traversal then exit")
		}
	}
	return nil
}

func snapshotDescription(snapshot browserSnapshot) string {
	return strings.TrimSpace(snapshot.ActiveRole + " " + snapshot.ActiveName)
}

func voiceOverPhrase(raw []byte, name string, action captureActionRecord) (string, error) {
	events, err := parseVoiceOverLogEvents(raw)
	if err != nil {
		return "", err
	}
	actionIssued, err := captureRFC3339Timestamp(action.ActionIssuedAt, "action issued")
	if err != nil {
		return "", err
	}
	logEnded, err := captureRFC3339Timestamp(action.LogEndedAt, "log end")
	if err != nil {
		return "", err
	}
	for _, event := range events {
		timestamp, timestampErr := captureRFC3339Timestamp(event.Timestamp, "VoiceOver event")
		if timestampErr != nil || !timestamp.After(actionIssued) || timestamp.After(logEnded) || event.ProcessID != action.VoiceOverPID || event.Subsystem != action.VoiceOverSubsystem {
			continue
		}
		message := strings.TrimSpace(event.EventMessage)
		lower := strings.ToLower(message)
		if strings.Contains(lower, "voiceover") && strings.Contains(lower, "region") && strings.Contains(lower, strings.ToLower(name)) {
			return message, nil
		}
	}
	return "", fmt.Errorf("raw action-bound VoiceOver JSON lacks named-region event inside exact action interval")
}

func parseVoiceOverLogEvents(raw []byte) ([]voiceOverLogEvent, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return nil, fmt.Errorf("VoiceOver log is empty")
	}
	var events []voiceOverLogEvent
	if err := json.Unmarshal(trimmed, &events); err == nil {
		if len(events) == 0 {
			return nil, fmt.Errorf("VoiceOver JSON log has no events")
		}
		return events, nil
	}
	var wrapped struct {
		Events []voiceOverLogEvent `json:"events"`
	}
	if err := json.Unmarshal(trimmed, &wrapped); err == nil && len(wrapped.Events) > 0 {
		return wrapped.Events, nil
	}
	return nil, fmt.Errorf("VoiceOver output must be JSON event bytes, not cumulative prose")
}

func meaningfulCaptureText(value string) bool {
	trimmed := strings.TrimSpace(value)
	return trimmed == value && len(trimmed) >= 3 && !strings.Contains(strings.ToLower(trimmed), "placeholder")
}

type browserCaptureScriptConfig struct {
	Schema         string `json:"schema"`
	Pair           string `json:"pair"`
	State          string `json:"state,omitempty"`
	Command        string `json:"command,omitempty"`
	Route          string `json:"route"`
	QueryKey       string `json:"queryKey"`
	Challenge      string `json:"challenge"`
	CandidateTree  string `json:"candidateTree"`
	ManifestSHA256 string `json:"manifestSHA256"`
	Selector       string `json:"selector"`
	Setup          string `json:"setup,omitempty"`
	Phase          string `json:"phase,omitempty"`
	ActionToken    string `json:"actionToken,omitempty"`
}

func browserPrepareScript(plan capturePlan, state captureState, identity candidateIdentity, challenge captureChallenge) string {
	selector := "#scroll-region-default [data-goshtoso-scroll-viewport]"
	if state.Name == "no-overflow" {
		selector = "#scroll-region-no-overflow [data-goshtoso-scroll-viewport]"
	}
	setup := "blur"
	switch state.Name {
	case "no-overflow", "focused":
		setup = "preceding-focus"
	case "start":
		// Home must cause the target start boundary. Set up a real non-start
		// boundary first; otherwise the final Home could be a no-op claim.
		setup = "middle"
	case "middle":
		// PageDown must move from start into middle, not merely observe a
		// state directly established by adapter preparation.
		setup = "start"
	case "end":
		setup = "middle"
	}
	return browserCaptureScript(plan, browserPrepareSource, browserCaptureScriptConfig{
		Schema: captureBrowserStateSchema, Pair: plan.Pair, State: state.Name, Command: state.Command, Route: captureRoute, QueryKey: captureServerQuery, Challenge: challenge.Challenge, CandidateTree: identity.CandidateTree,
		ManifestSHA256: identity.ManifestSHA256, Selector: selector, Setup: setup, ActionToken: captureActionToken(challenge.Challenge, state.Name),
	})
}

func browserReadStateScript(plan capturePlan, state captureState, phase string, identity candidateIdentity, challenge captureChallenge) string {
	selector := "#scroll-region-default [data-goshtoso-scroll-viewport]"
	if state.Name == "no-overflow" {
		selector = "#scroll-region-no-overflow [data-goshtoso-scroll-viewport]"
	}
	return browserCaptureScript(plan, browserReadStateSource, browserCaptureScriptConfig{
		Schema: captureBrowserStateSchema, Pair: plan.Pair, State: state.Name, Command: state.Command, Route: captureRoute,
		QueryKey: captureServerQuery, Challenge: challenge.Challenge, CandidateTree: identity.CandidateTree, ManifestSHA256: identity.ManifestSHA256,
		Selector: selector, Phase: phase, ActionToken: captureActionToken(challenge.Challenge, state.Name),
	})
}

func browserWindowScript(plan capturePlan, identity candidateIdentity, challenge captureChallenge) string {
	return browserCaptureScript(plan, browserWindowSource, browserCaptureScriptConfig{
		Schema: captureBrowserStateSchema, Pair: plan.Pair, Route: captureRoute, QueryKey: captureServerQuery,
		Challenge: challenge.Challenge, CandidateTree: identity.CandidateTree, ManifestSHA256: identity.ManifestSHA256,
	})
}

func browserCaptureScript(plan capturePlan, source string, config browserCaptureScriptConfig) string {
	encoded, err := json.Marshal(config)
	if err != nil {
		panic(fmt.Sprintf("encode fixed browser capture configuration: %v", err))
	}
	javaScript := strings.Replace(source, browserCaptureConfigMarker, string(encoded), 1)
	if strings.Contains(javaScript, browserCaptureConfigMarker) {
		panic("embedded browser capture program has an unresolved configuration marker")
	}
	return browserJavaScriptScript(plan, javaScript)
}

func browserJavaScriptScript(plan capturePlan, javaScript string) string {
	quoted := strconv.Quote(javaScript)
	if plan.BrowserApp == "Safari" {
		return fmt.Sprintf("tell application %q\n  activate\n  delay 0.25\n  return do JavaScript %s in current tab of front window\nend tell", plan.BrowserApp, quoted)
	}
	return fmt.Sprintf("tell application %q\n  activate\n  delay 0.25\n  return execute javascript %s in active tab of front window\nend tell", plan.BrowserApp, quoted)
}

func browserKeyboardScript(plan capturePlan, keyCode int) string {
	return fmt.Sprintf("tell application %q to activate\ndelay 0.25\ntell application \"System Events\"\n  tell process %q\n    set frontmost to true\n    key code %d\n  end tell\nend tell", plan.BrowserApp, plan.BrowserApp, keyCode)
}

func browserNavigateScript(plan capturePlan, target string) string {
	encoded, err := json.Marshal(target)
	if err != nil {
		panic(fmt.Sprintf("encode direct browser navigation target: %v", err))
	}
	return browserJavaScriptScript(plan, "window.location.assign("+string(encoded)+");")
}
