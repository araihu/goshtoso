package fileinput

import (
	"bytes"
	"context"
	"slices"
	"strings"
	"testing"

	"github.com/a-h/templ"
)

// renderConfig renders FileInput(cfg) to a string for assertions.
func renderConfig(t *testing.T, cfg Config) string {
	t.Helper()
	var buf bytes.Buffer
	if err := FileInput(cfg).Render(context.Background(), &buf); err != nil {
		t.Fatalf("render fileinput: %v", err)
	}
	return buf.String()
}

func TestCoverageRenderDefaultFileinput(t *testing.T) {
	html := renderConfig(t, Config{})
	for _, want := range []string{
		"class=",
		`data-fileinput-variant="dropzone"`,
		`type="file"`,
		"border-dashed",
		"border-control-outline dark:border-control-outline-dark",
		"Browse",
		"or drag and drop here",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("default render missing %q in %s", want, html)
		}
	}
	// A bare default config should not emit accept/required/disabled/aria-describedby.
	for _, notWant := range []string{"accept=", "required", "disabled", "aria-describedby"} {
		if strings.Contains(html, notWant) {
			t.Fatalf("default render unexpectedly contains %q in %s", notWant, html)
		}
	}
}

func TestCoverageDropZoneAllBranches(t *testing.T) {
	html := renderConfig(t, Config{
		ID:         "cover",
		Name:       "cover",
		Label:      "Cover Picture",
		Accept:     "image/*",
		HelperText: "PNG, JPG - Max 5MB",
		Required:   true,
		InputAttrs: templ.Attributes{"hx-post": "/upload"},
	})

	for _, want := range []string{
		"<span", "Cover Picture", // Label branch
		`accept="image/*"`,                // Accept branch
		"required",                        // Required branch
		`aria-describedby="cover-helper"`, // HelperText -> describedby branch
		`id="cover-helper"`,               // helper small branch
		"PNG, JPG - Max 5MB",
		`hx-post="/upload"`, // InputAttrs spread
		`for="cover"`,
		`id="cover"`,
		`name="cover"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("dropzone render missing %q in %s", want, html)
		}
	}
	if strings.Contains(html, "disabled") {
		t.Fatalf("non-disabled dropzone should not contain disabled attr: %s", html)
	}
}

func TestCoverageDropZoneDisabled(t *testing.T) {
	html := renderConfig(t, Config{ID: "x", Disabled: true})
	for _, want := range []string{"disabled", "opacity-50", "cursor-not-allowed"} {
		if !strings.Contains(html, want) {
			t.Fatalf("disabled dropzone missing %q in %s", want, html)
		}
	}
}

func TestCoverageUploadAppearanceAllBranches(t *testing.T) {
	html := renderConfig(t, Config{
		Appearance: AppearanceUpload,
		ID:         "resume",
		Name:       "resume",
		Label:      "Resume",
		Accept:     ".pdf,.docx",
		HelperText: "PDF or DOCX",
		Required:   true,
		InputAttrs: templ.Attributes{"data-test": "upload"},
	})

	for _, want := range []string{
		`data-fileinput-variant="upload"`,
		"No file selected",
		"rounded-radius",
		"bg-surface-alt",
		"Resume",
		`accept=".pdf,.docx"`,
		"required",
		`aria-describedby="resume-helper"`,
		`id="resume-helper"`,
		"PDF or DOCX",
		`data-test="upload"`,
		`for="resume"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("upload render missing %q in %s", want, html)
		}
	}
	// Upload variant must not render the drop zone.
	if strings.Contains(html, "border-dashed") {
		t.Fatalf("upload variant should not render drop zone: %s", html)
	}
}

func TestCoverageUploadDisabled(t *testing.T) {
	html := renderConfig(t, Config{Appearance: AppearanceUpload, ID: "u", Disabled: true})
	for _, want := range []string{"disabled", "cursor-not-allowed", "opacity-75"} {
		if !strings.Contains(html, want) {
			t.Fatalf("disabled upload missing %q in %s", want, html)
		}
	}
}

func TestDisabledFileInputsDimContentWithoutFadingBoundaries(t *testing.T) {
	dropzone := Config{Disabled: true}.dropZoneClasses()
	if slicesContain(strings.Fields(dropzone), "opacity-50") {
		t.Fatalf("dropzone boundary must not receive direct disabled opacity: %q", dropzone)
	}
	if !strings.Contains(dropzone, "[&>*]:opacity-50") {
		t.Fatalf("dropzone content must retain scoped disabled opacity: %q", dropzone)
	}

	upload := Config{Appearance: AppearanceUpload, Disabled: true}
	if slicesContain(strings.Fields(upload.uploadControlClasses()), "opacity-75") {
		t.Fatalf("upload boundary must not receive direct disabled opacity: %q", upload.uploadControlClasses())
	}
	if !strings.Contains(upload.uploadFileNameClasses(), "opacity-75") || !strings.Contains(upload.uploadButtonClasses(), "opacity-75") {
		t.Fatalf("upload content must retain disabled opacity: filename=%q button=%q", upload.uploadFileNameClasses(), upload.uploadButtonClasses())
	}
}

func slicesContain(values []string, want string) bool {
	return slices.Contains(values, want)
}

func TestCoverageNoLabelNoHelper(t *testing.T) {
	// Drop zone without label/helper exercises the false branches.
	html := renderConfig(t, Config{ID: "bare"})
	if strings.Contains(html, "<span class=") && strings.Contains(html, "pl-0.5 text-sm") {
		t.Fatalf("bare dropzone should not render label span: %s", html)
	}
	if strings.Contains(html, "aria-describedby") {
		t.Fatalf("bare dropzone should not set aria-describedby: %s", html)
	}

	// Upload variant without label/helper.
	up := renderConfig(t, Config{Appearance: AppearanceUpload, ID: "bareup"})
	if strings.Contains(up, "aria-describedby") {
		t.Fatalf("bare upload should not set aria-describedby: %s", up)
	}
}

func TestCoverageIsUpload(t *testing.T) {
	if (Config{}).isUpload() {
		t.Fatal("default config should not be upload")
	}
	if !(Config{Appearance: AppearanceUpload}).isUpload() {
		t.Fatal("upload variant should report isUpload")
	}
}

func TestCoverageContainerClasses(t *testing.T) {
	base := Config{}.containerClasses()
	if !strings.Contains(base, "flex w-full") {
		t.Fatalf("unexpected container base: %q", base)
	}
	withRoot := Config{RootClass: "mt-4"}.containerClasses()
	if !strings.HasSuffix(withRoot, " mt-4") {
		t.Fatalf("RootClass not appended to container: %q", withRoot)
	}
}

func TestCoverageUploadContainerClasses(t *testing.T) {
	base := Config{}.uploadContainerClasses()
	if !strings.Contains(base, "text-left") {
		t.Fatalf("unexpected upload container base: %q", base)
	}
	withRoot := Config{RootClass: "gap-3"}.uploadContainerClasses()
	if !strings.HasSuffix(withRoot, " gap-3") {
		t.Fatalf("RootClass not appended to upload container: %q", withRoot)
	}
}

func TestCoverageClassHelpers(t *testing.T) {
	cfg := Config{}
	checks := map[string]string{
		"labelClasses":          cfg.labelClasses(),
		"browseLabelClasses":    cfg.browseLabelClasses(),
		"uploadFileNameClasses": cfg.uploadFileNameClasses(),
		"helperTextClasses":     cfg.helperTextClasses(),
	}
	for name, got := range checks {
		if strings.TrimSpace(got) == "" {
			t.Fatalf("%s returned empty", name)
		}
	}
	if !strings.Contains(cfg.helperTextClasses(), "text-on-surface-muted") ||
		!strings.Contains(cfg.helperTextClasses(), "dark:text-on-surface-dark-muted") {
		t.Fatalf("helper text must use semantic muted tokens: %q", cfg.helperTextClasses())
	}
	if strings.Contains(cfg.helperTextClasses(), "text-on-surface/60") {
		t.Fatalf("helper text retained opacity hierarchy: %q", cfg.helperTextClasses())
	}

	if !strings.Contains(cfg.dropZoneClasses(), "border-dashed") {
		t.Fatal("enabled dropZoneClasses missing border-dashed")
	}
	if !strings.Contains((Config{Disabled: true}).dropZoneClasses(), "opacity-50") {
		t.Fatal("disabled dropZoneClasses missing opacity-50")
	}

	if !strings.Contains(cfg.uploadControlClasses(), "cursor-pointer") {
		t.Fatal("enabled uploadControlClasses missing cursor-pointer")
	}
	if !strings.Contains(cfg.uploadControlClasses(), "border-control-outline") ||
		!strings.Contains(cfg.uploadControlClasses(), "dark:border-control-outline-dark") {
		t.Fatalf("uploadControlClasses missing control boundary roles: %q", cfg.uploadControlClasses())
	}
	if !strings.Contains((Config{Disabled: true}).uploadControlClasses(), "cursor-not-allowed") {
		t.Fatal("disabled uploadControlClasses missing cursor-not-allowed")
	}

	if strings.Contains(cfg.uploadButtonClasses(), "cursor-not-allowed") {
		t.Fatal("enabled uploadButtonClasses should not be cursor-not-allowed")
	}
	if !strings.Contains((Config{Disabled: true}).uploadButtonClasses(), "cursor-not-allowed") {
		t.Fatal("disabled uploadButtonClasses missing cursor-not-allowed")
	}
}
