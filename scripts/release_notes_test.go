package scripts_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestReleaseNotes(t *testing.T) {
	const heading = "## [v1.2.3] - 2026-11-05\n"
	for _, tc := range []struct {
		name, tag, changelog, want, problem string
	}{
		{"dated section", "v1.2.3", heading + "\n### Fixed\n\n- Fixed dragging.\n\n## [v1.2.2] - 2026-11-01\nOld notes.\n", "### Fixed\n\n- Fixed dragging.\n", ""},
		{"exact version", "v1.2.3", "## [v1.2.30] - 2026-11-05\nOther version.\n", "", "missing"},
		{"missing section", "v1.2.3", "# Changelog\n", "", "missing"},
		{"missing date", "v1.2.3", "## [v1.2.3]\nNotes.\n", "", "dated heading"},
		{"malformed date", "v1.2.3", "## [v1.2.3] - tomorrow\nNotes.\n", "", "dated heading"},
		{"empty section", "v1.2.3", heading + " \n### Fixed\n## [v1.2.2] - 2026-11-01\nOld notes.\n", "", "empty"},
		{"comments only", "v1.2.3", heading + "<!--\nTODO\n-->\n", "", "empty"},
		{"duplicate section", "v1.2.3", heading + "First.\n" + heading + "Second.\n", "", "duplicate"},
		{"next ordinary heading", "v1.2.3", heading + "Notes.\n## Support\nNot release notes.\n", "Notes.\n", ""},
		{"fenced heading", "v1.2.3", "````md\n```\n" + heading + "Example.\n````\n", "", "missing"},
		{"indented fence", "v1.2.3", "   ~~~md\n" + heading + "Example.\n   ~~~\n", "", "missing"},
		{"commented heading", "v1.2.3", "<!--\n" + heading + "Example.\n-->\n", "", "missing"},
		{"fenced content", "v1.2.3", heading + "Notes.\n```md\n" + heading + "Example.\n```\n", "Notes.\n```md\n" + heading + "Example.\n```\n", ""},
		{"CRLF", "v1.2.3", strings.ReplaceAll(heading+"Notes.\n", "\n", "\r\n"), "Notes.\n", ""},
		{"invalid tag", "v01.2.3", heading + "Notes.\n", "", "invalid stable release tag"},
		{"prerelease tag", "v1.2.3-rc.1", heading + "Notes.\n", "", "invalid stable release tag"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "CHANGELOG.md")
			if err := os.WriteFile(path, []byte(tc.changelog), 0o600); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command("bash", "release-notes", tc.tag, path)
			var stderr strings.Builder
			cmd.Stderr = &stderr
			output, err := cmd.Output()
			if tc.problem != "" {
				if err == nil || len(output) != 0 || !strings.Contains(stderr.String(), tc.problem) {
					t.Fatalf("stdout=%q, stderr=%q, err=%v; want rejection containing %q", output, stderr.String(), err, tc.problem)
				}
				return
			}
			if err != nil || string(output) != tc.want {
				t.Fatalf("stdout=%q, stderr=%q, err=%v; want %q", output, stderr.String(), err, tc.want)
			}
		})
	}
}

func TestReleasePublisherExtractsSharedNotes(t *testing.T) {
	module, err := os.ReadFile("../.dagger/src/index.ts")
	if err != nil {
		t.Fatal(err)
	}
	command := regexp.MustCompile(`(?m)^scripts/release-notes "\$TAG" > /tmp/release-notes.md$`).FindString(string(module))
	if command == "" {
		t.Fatal("publisher must extract notes through the shared validator")
	}
	for _, present := range []bool{true, false} {
		dir := t.TempDir()
		if err := os.Mkdir(filepath.Join(dir, "scripts"), 0o700); err != nil {
			t.Fatal(err)
		}
		parser, err := os.ReadFile("release-notes")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "scripts/release-notes"), parser, 0o700); err != nil {
			t.Fatal(err)
		}
		notes := "# Changelog\n"
		if present {
			notes += "## [v1.2.3] - 2026-11-05\nFixed release validation.\n"
		}
		if err := os.WriteFile(filepath.Join(dir, "CHANGELOG.md"), []byte(notes), 0o600); err != nil {
			t.Fatal(err)
		}
		// Keep the publisher output inside the fixture rather than the host /tmp.
		cmd := exec.Command("bash", "-euc", "TAG=v1.2.3\n"+strings.ReplaceAll(command, "/tmp/release-notes.md", "release-notes.md"))
		cmd.Dir = dir
		output, err := cmd.CombinedOutput()
		if present {
			actual, readErr := os.ReadFile(filepath.Join(dir, "release-notes.md"))
			if err != nil || readErr != nil || string(actual) != "Fixed release validation.\n" {
				t.Fatalf("publisher notes=%q, err=%v, readErr=%v; output=%s", actual, err, readErr, output)
			}
		} else if err == nil {
			t.Fatal("publisher accepted missing notes")
		}
	}
}

func TestReleaseTagGuards(t *testing.T) {
	const tag = "v1.2.3"
	for _, scenario := range []string{"valid", "missing notes", "uncommitted notes", "dirty source", "untracked source", "existing tag"} {
		t.Run(scenario, func(t *testing.T) {
			dir := t.TempDir()
			runGit := func(args ...string) string {
				t.Helper()
				return strings.TrimSpace(runGitOutput(t, dir, append([]string{"-c", "core.hooksPath=/dev/null", "-c", "commit.gpgSign=false", "-c", "tag.gpgSign=false"}, args...)...))
			}
			runGit("init", "-q")
			runGit("config", "user.name", "Release test")
			runGit("config", "user.email", "release@example.invalid")
			runGit("config", "tag.gpgSign", "false")
			if err := os.Mkdir(filepath.Join(dir, "scripts"), 0o700); err != nil {
				t.Fatal(err)
			}
			for _, name := range []string{"release-notes", "release-tag"} {
				data, err := os.ReadFile(name)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, "scripts", name), data, 0o700); err != nil {
					t.Fatal(err)
				}
			}
			notes := "# Changelog\n"
			if scenario != "missing notes" && scenario != "uncommitted notes" {
				notes += "## [" + tag + "] - 2026-11-05\nFixed release validation.\n"
			}
			writeFile(t, filepath.Join(dir, "CHANGELOG.md"), notes)
			writeFile(t, filepath.Join(dir, "source.txt"), "committed\n")
			runGit("add", ".")
			runGit("commit", "-qm", "fixture")
			commit := runGit("rev-parse", "HEAD")
			switch scenario {
			case "uncommitted notes":
				writeFile(t, filepath.Join(dir, "CHANGELOG.md"), notes+"## ["+tag+"] - 2026-11-05\nUncommitted notes.\n")
			case "dirty source":
				writeFile(t, filepath.Join(dir, "source.txt"), "uncommitted\n")
			case "untracked source":
				writeFile(t, filepath.Join(dir, "new.txt"), "untracked\n")
			case "existing tag":
				runGit("tag", tag)
			}
			cmd := exec.Command("bash", filepath.Join(dir, "scripts", "release-tag"), tag)
			cmd.Dir = t.TempDir() // The helper must tag its own checkout, not the caller directory.
			output, err := cmd.CombinedOutput()
			if scenario == "valid" {
				if err != nil || runGit("rev-parse", "refs/tags/"+tag) != commit {
					t.Fatalf("valid release: %v\n%s", err, output)
				}
			} else if err == nil {
				t.Fatalf("invalid release created a tag: %s", output)
			}
			want := ""
			if scenario == "valid" || scenario == "existing tag" {
				want = tag
				if runGit("rev-parse", "refs/tags/"+tag) != commit {
					t.Fatal("existing tag changed")
				}
			}
			if tags := runGit("tag", "--list"); tags != want {
				t.Fatalf("tags=%q; want %q", tags, want)
			}
		})
	}
}
