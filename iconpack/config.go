package iconpack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"go.yaml.in/yaml/v3"
)

const iconpackConfigSchemaVersion = 1

// Config is the Goshtoso-owned .iconpack.yaml contract. Its lock is always
// the sibling .iconpack.lock.yaml (or IconpackLockPath when explicitly set);
// Muamba manifests are never discovered or modified.
type Config struct {
	SchemaVersion int            `json:"schemaVersion" yaml:"schemaVersion"`
	Sources       []ConfigSource `json:"sources" yaml:"sources"`
}

// ConfigSource describes a GitHub tree/archive, a single remote file, or a
// source that is resolved by the iconpack acquisition adapter.
type ConfigSource struct {
	ID              string   `json:"id" yaml:"id"`
	URL             string   `json:"url" yaml:"url"`
	Kind            string   `json:"kind,omitempty" yaml:"kind,omitempty"`
	PackName        string   `json:"packName,omitempty" yaml:"packName,omitempty"`
	Path            string   `json:"path,omitempty" yaml:"path,omitempty"`
	Paths           []string `json:"paths,omitempty" yaml:"paths,omitempty"`
	License         string   `json:"license" yaml:"license"`
	LicensePath     string   `json:"licensePath" yaml:"licensePath"`
	LicenseURL      string   `json:"licenseUrl,omitempty" yaml:"licenseUrl,omitempty"`
	StripComponents int      `json:"stripComponents,omitempty" yaml:"stripComponents,omitempty"`
	MaxFiles        int      `json:"maxFiles,omitempty" yaml:"maxFiles,omitempty"`
	MaxUnpackedSize string   `json:"maxUnpackedSize,omitempty" yaml:"maxUnpackedSize,omitempty"`
}

type backendManifest struct {
	Schema    int                        `yaml:"schema"`
	Resources map[string]backendResource `yaml:"resources"`
}

type backendResource struct {
	Version     string                      `yaml:"version"`
	Downloads   map[string]backendDownload  `yaml:"downloads,omitempty"`
	Directories map[string]backendDirectory `yaml:"directories,omitempty"`
}

type backendDownload struct {
	URL      string `yaml:"url"`
	Path     string `yaml:"path"`
	MaxSize  string `yaml:"max_size,omitempty"`
	Platform string `yaml:"platform,omitempty"`
}

type backendDirectory struct {
	URL             string   `yaml:"url"`
	Archive         string   `yaml:"archive"`
	Path            string   `yaml:"path"`
	Include         []string `yaml:"include"`
	Exclude         []string `yaml:"exclude,omitempty"`
	StripComponents int      `yaml:"strip_components,omitempty"`
	MaxSize         string   `yaml:"max_size,omitempty"`
	MaxFiles        int      `yaml:"max_files"`
	MaxUnpackedSize string   `yaml:"max_unpacked_size"`
}

func readIconpackConfig(filename string) ([]byte, Config, error) {
	contents, err := readExternalManifest(filename)
	if err != nil {
		return nil, Config{}, fmt.Errorf("read iconpack config: %w", err)
	}
	config, err := decodeIconpackConfig(filename, contents)
	if err != nil {
		return nil, Config{}, err
	}
	return contents, config, nil
}

func decodeIconpackConfig(filename string, contents []byte) (Config, error) {
	var config Config
	trimmed := bytes.TrimSpace(contents)
	if filepath.Ext(filename) == ".json" || len(trimmed) > 0 && trimmed[0] == '{' {
		decoder := json.NewDecoder(bytes.NewReader(contents))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&config); err != nil {
			return Config{}, fmt.Errorf("decode JSON iconpack config: %w", err)
		}
		if err := requireJSONEOF(decoder, "iconpack config"); err != nil {
			return Config{}, err
		}
	} else {
		decoder := yaml.NewDecoder(bytes.NewReader(contents))
		decoder.KnownFields(true)
		if err := decoder.Decode(&config); err != nil {
			return Config{}, fmt.Errorf("decode YAML iconpack config: %w", err)
		}
		var extra any
		if err := decoder.Decode(&extra); err != io.EOF {
			if err == nil {
				return Config{}, fmt.Errorf("iconpack config contains more than one YAML document")
			}
			return Config{}, fmt.Errorf("decode YAML iconpack config: %w", err)
		}
	}
	if err := validateIconpackConfig(config); err != nil {
		return Config{}, err
	}
	return config, nil
}

func validateIconpackConfig(config Config) error {
	if config.SchemaVersion != iconpackConfigSchemaVersion {
		return fmt.Errorf("unsupported iconpack schemaVersion %d: want %d", config.SchemaVersion, iconpackConfigSchemaVersion)
	}
	if len(config.Sources) == 0 {
		return fmt.Errorf("iconpack sources must not be empty")
	}
	seen := make(map[string]struct{}, len(config.Sources))
	for index, source := range config.Sources {
		if err := validateConfigSource(index, source); err != nil {
			return err
		}
		if _, exists := seen[source.ID]; exists {
			return fmt.Errorf("iconpack source id %q is duplicated", source.ID)
		}
		seen[source.ID] = struct{}{}
	}
	return nil
}

func validateConfigSource(index int, source ConfigSource) error {
	if !validResourceName(source.ID) {
		return fmt.Errorf("iconpack source %d id %q must be lower-kebab", index, source.ID)
	}
	if strings.TrimSpace(source.URL) == "" || strings.ContainsAny(source.URL, "\r\n\x00") {
		return fmt.Errorf("iconpack source %q url is required", source.ID)
	}
	if strings.TrimSpace(source.License) == "" || strings.ContainsAny(source.License, "\r\n\x00") {
		return fmt.Errorf("iconpack source %q license is required", source.ID)
	}
	if err := safeRelativePath(source.LicensePath); err != nil {
		return fmt.Errorf("iconpack source %q licensePath: %w", source.ID, err)
	}
	fileSource := source.Kind == "file" || strings.HasSuffix(strings.ToLower(source.URL), ".svg")
	if source.LicenseURL == "" && fileSource {
		return fmt.Errorf("iconpack file source %q requires licenseUrl", source.ID)
	}
	if source.Kind == "file" && source.Path != "" {
		if err := safeRelativePath(source.Path); err != nil {
			return fmt.Errorf("iconpack source %q path: %w", source.ID, err)
		}
		if !strings.EqualFold(filepath.Ext(source.Path), ".svg") {
			return fmt.Errorf("iconpack source %q path %q is not an SVG", source.ID, source.Path)
		}
	}
	if source.StripComponents < 0 {
		return fmt.Errorf("iconpack source %q stripComponents must not be negative", source.ID)
	}
	if source.MaxFiles < 0 || source.MaxFiles > maxArchiveFiles {
		return fmt.Errorf("iconpack source %q maxFiles must be between 0 and %d", source.ID, maxArchiveFiles)
	}
	for _, item := range source.Paths {
		if err := safeRelativePath(item); err != nil {
			return fmt.Errorf("iconpack source %q path %q: %w", source.ID, item, err)
		}
		if !strings.EqualFold(filepath.Ext(item), ".svg") {
			return fmt.Errorf("iconpack source %q selected path %q is not an SVG", source.ID, item)
		}
	}
	return nil
}

func validResourceName(value string) bool {
	if value == "" || value[0] < 'a' || value[0] > 'z' {
		return false
	}
	for _, character := range value {
		if character >= 'a' && character <= 'z' || character >= '0' && character <= '9' || character == '-' {
			continue
		}
		return false
	}
	return !strings.HasSuffix(value, "-") && !strings.Contains(value, "--")
}

func isGitTreeURL(raw string) bool {
	parsed, err := url.Parse(raw)
	return err == nil && parsed.Host == "github.com" && strings.Contains("/"+parsed.Path, "/tree/")
}

type resolvedConfigSource struct {
	ConfigSource
	PackName      string
	Version       string
	ArchiveURL    string
	RootSegments  []string
	Archive       string
	Kind          string
	IconPath      string
	LicenseTarget string
}

func resolveConfigSource(source ConfigSource) (resolvedConfigSource, error) {
	result := resolvedConfigSource{ConfigSource: source, PackName: normalizeWebPath(source.PackName), Archive: "tar.gz", Kind: source.Kind}
	if source.PackName != "" && result.PackName == "" {
		return resolvedConfigSource{}, fmt.Errorf("iconpack source %q has an empty normalized packName", source.ID)
	}
	if isGitTreeURL(source.URL) || source.Kind == "git-tree" {
		archiveURL, version, root, repo, err := parseGitHubTreeURL(source.URL)
		if err != nil {
			return resolvedConfigSource{}, err
		}
		if source.PackName == "" {
			result.PackName = normalizeWebPath(repo)
		}
		if result.PackName == "" {
			return resolvedConfigSource{}, fmt.Errorf("iconpack source %q has an empty repository name", source.ID)
		}
		result.ArchiveURL, result.Version, result.RootSegments = archiveURL, version, root
		result.Kind = "git-tree"
		result.LicenseTarget = filepath.ToSlash(filepath.Join(source.ID, source.LicensePath))
		return result, nil
	}
	if source.Kind == "file" || strings.HasSuffix(strings.ToLower(source.URL), ".svg") {
		result.Kind = "file"
		result.Version = "source"
		result.IconPath = source.Path
		if result.IconPath == "" {
			parsed, _ := url.Parse(source.URL)
			result.IconPath = path.Base(parsed.Path)
		}
		result.LicenseTarget = filepath.ToSlash(filepath.Join(source.ID, source.LicensePath))
		return result, nil
	}
	if source.Kind != "archive" {
		return resolvedConfigSource{}, fmt.Errorf("iconpack source %q must be a GitHub tree, file, or archive", source.ID)
	}
	result.ArchiveURL, result.Version = source.URL, "source"
	result.RootSegments = strings.Split(strings.Trim(source.PackName, "/"), "/")
	result.Kind = "archive"
	result.LicenseTarget = filepath.ToSlash(filepath.Join(source.ID, source.LicensePath))
	return result, nil
}

func parseGitHubTreeURL(raw string) (archiveURL, version string, root []string, repo string, err error) {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme != "https" || parsed.Host != "github.com" {
		return "", "", nil, "", fmt.Errorf("iconpack GitHub tree URL %q must be an HTTPS github.com URL", raw)
	}
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) < 4 || parts[2] != "tree" || parts[0] == "" || parts[1] == "" || parts[3] == "" {
		return "", "", nil, "", fmt.Errorf("iconpack GitHub tree URL %q must have /owner/repository/tree/ref[/root]", raw)
	}
	owner, repository, ref := parts[0], parts[1], parts[3]
	root = append([]string(nil), parts[4:]...)
	archiveURL = fmt.Sprintf("https://codeload.github.com/%s/%s/tar.gz/%s", owner, repository, ref)
	return archiveURL, ref, root, repository, nil
}

func normalizeWebPath(value string) string {
	value = strings.TrimSuffix(strings.ReplaceAll(value, "\\", "/"), filepath.Ext(value))
	parts := strings.Split(value, "/")
	var output []string
	for _, part := range parts {
		var builder strings.Builder
		for _, character := range strings.ToLower(part) {
			if character >= 'a' && character <= 'z' || character >= '0' && character <= '9' {
				builder.WriteRune(character)
			} else if builder.Len() > 0 {
				builder.WriteByte('-')
			}
		}
		if item := strings.Trim(builder.String(), "-"); item != "" {
			output = append(output, item)
		}
	}
	return strings.Join(output, "-")
}

func backendForConfig(config Config) (backendManifest, []resolvedConfigSource, error) {
	backend := backendManifest{Schema: 1, Resources: make(map[string]backendResource, len(config.Sources))}
	resolved := make([]resolvedConfigSource, 0, len(config.Sources))
	for _, declared := range config.Sources {
		source, err := resolveConfigSource(declared)
		if err != nil {
			return backendManifest{}, nil, err
		}
		resolved = append(resolved, source)
		resource := backendResource{Version: source.Version}
		base := filepath.ToSlash(filepath.Join(".iconpack", "sources", source.ID))
		if source.Kind == "file" {
			resource.Downloads = map[string]backendDownload{
				"icon":    {URL: source.URL, Path: filepath.ToSlash(filepath.Join(base, source.IconPath)), MaxSize: "64MiB"},
				"license": {URL: source.LicenseURL, Path: filepath.ToSlash(filepath.Join(base, source.LicensePath)), MaxSize: "16MiB"},
			}
		} else {
			strip := source.StripComponents
			if source.Kind == "git-tree" {
				strip = 1 + len(source.RootSegments)
			}
			maxFiles := source.MaxFiles
			if maxFiles == 0 {
				maxFiles = maxArchiveFiles
			}
			maxUnpacked := source.MaxUnpackedSize
			if maxUnpacked == "" {
				maxUnpacked = "1GiB"
			}
			resource.Directories = map[string]backendDirectory{
				"tree": {URL: source.ArchiveURL, Archive: source.Archive, Path: base, Include: []string{"**"}, StripComponents: strip, MaxSize: "256MiB", MaxFiles: maxFiles, MaxUnpackedSize: maxUnpacked},
			}
			if source.LicenseURL != "" {
				resource.Downloads = map[string]backendDownload{
					"license": {URL: source.LicenseURL, Path: filepath.ToSlash(filepath.Join(base, source.LicensePath)), MaxSize: "16MiB"},
				}
			}
		}
		backend.Resources[declared.ID] = resource
	}
	return backend, resolved, nil
}

func marshalBackendManifest(manifest backendManifest) ([]byte, error) {
	var out bytes.Buffer
	encoder := yaml.NewEncoder(&out)
	encoder.SetIndent(2)
	if err := encoder.Encode(manifest); err != nil {
		return nil, err
	}
	if err := encoder.Close(); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func sortedConfigSources(sources []resolvedConfigSource) []resolvedConfigSource {
	result := append([]resolvedConfigSource(nil), sources...)
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}
