package assets

import (
	"net/http"
	"strings"
)

const (
	// ImmutableCacheControl is reserved for URLs whose path identifies exact
	// bytes through a version or content hash.
	ImmutableCacheControl = "public, max-age=31536000, immutable"
	// RevalidateCacheControl keeps mutable aliases cacheable while requiring a
	// freshness check before reuse.
	RevalidateCacheControl = "public, max-age=0, must-revalidate"
)

// CacheControl returns the HTTP Cache-Control value for one asset path.
//
// Exact semantic-version runtime and license paths, numeric chart-control
// generations, and paths containing a content-hash token are immutable. Every
// other path is a mutable alias and must revalidate. Query parameters do not
// establish identity: callers must put version or content identity in the path
// before this function permits immutable caching.
func CacheControl(assetPath string) string {
	segments, ok := cachePathSegments(assetPath)
	if !ok {
		return RevalidateCacheControl
	}
	if hasExactVersionedAssetPath(segments) || hasContentAddressedPath(segments) {
		return ImmutableCacheControl
	}
	return RevalidateCacheControl
}

// WithCacheControl applies CacheControl to responses from next. It enforces the
// shared policy when wrapped handlers set their own Cache-Control header.
func WithCacheControl(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		cacheControl := CacheControl(request.URL.Path)
		writer.Header().Set("Cache-Control", cacheControl)
		next.ServeHTTP(&cacheControlResponseWriter{
			ResponseWriter: writer,
			cacheControl:   cacheControl,
		}, request)
	})
}

type cacheControlResponseWriter struct {
	http.ResponseWriter
	cacheControl string
	wroteHeader  bool
}

func (writer *cacheControlResponseWriter) WriteHeader(statusCode int) {
	if writer.wroteHeader {
		return
	}
	writer.wroteHeader = true
	cacheControl := writer.cacheControl
	if !cacheableAssetStatus(statusCode) {
		cacheControl = RevalidateCacheControl
	}
	writer.Header().Set("Cache-Control", cacheControl)
	writer.ResponseWriter.WriteHeader(statusCode)
}

func (writer *cacheControlResponseWriter) Write(data []byte) (int, error) {
	if !writer.wroteHeader {
		writer.WriteHeader(http.StatusOK)
	}
	return writer.ResponseWriter.Write(data)
}

func (writer *cacheControlResponseWriter) Unwrap() http.ResponseWriter {
	return writer.ResponseWriter
}

func cacheableAssetStatus(statusCode int) bool {
	return statusCode == http.StatusOK || statusCode == http.StatusPartialContent || statusCode == http.StatusNotModified
}

func cachePathSegments(assetPath string) ([]string, bool) {
	if assetPath == "" || strings.ContainsAny(assetPath, "?#\\") {
		return nil, false
	}
	trimmed := strings.Trim(assetPath, "/")
	if trimmed == "" {
		return nil, false
	}
	segments := strings.Split(trimmed, "/")
	for _, segment := range segments {
		if segment == "" || segment == "." || segment == ".." {
			return nil, false
		}
	}
	return segments, true
}

func hasExactVersionedAssetPath(segments []string) bool {
	for index := range segments {
		remaining := len(segments) - index
		if remaining >= 5 && segments[index] == "js" && segments[index+1] == "runtime" &&
			isExactSemanticVersion(segments[index+3]) {
			return true
		}
		if remaining >= 4 && segments[index] == "licenses" &&
			isExactSemanticVersion(segments[index+2]) {
			return true
		}
		if remaining >= 4 && segments[index] == "js" && segments[index+1] == "controls" &&
			isPositiveDecimal(segments[index+2]) {
			return true
		}
	}
	return false
}

func isExactSemanticVersion(value string) bool {
	value = strings.TrimPrefix(value, "v")
	coreAndPrerelease, build, hasBuild := strings.Cut(value, "+")
	if hasBuild && (strings.ContainsRune(build, '+') || !validVersionIdentifiers(build, false)) {
		return false
	}
	core, prerelease, hasPrerelease := strings.Cut(coreAndPrerelease, "-")
	if hasPrerelease && !validVersionIdentifiers(prerelease, true) {
		return false
	}
	parts := strings.Split(core, ".")
	return len(parts) == 3 && validCoreVersionPart(parts[0]) && validCoreVersionPart(parts[1]) && validCoreVersionPart(parts[2])
}

func validCoreVersionPart(value string) bool {
	return isDecimal(value) && (value == "0" || value[0] != '0')
}

func validVersionIdentifiers(value string, rejectLeadingZero bool) bool {
	for identifier := range strings.SplitSeq(value, ".") {
		if identifier == "" || (rejectLeadingZero && len(identifier) > 1 && identifier[0] == '0' && isDecimal(identifier)) {
			return false
		}
		for _, character := range identifier {
			if !isVersionIdentifierCharacter(character) {
				return false
			}
		}
	}
	return true
}

func isVersionIdentifierCharacter(character rune) bool {
	return (character >= '0' && character <= '9') ||
		(character >= 'a' && character <= 'z') ||
		(character >= 'A' && character <= 'Z') || character == '-'
}

func isPositiveDecimal(value string) bool {
	return isDecimal(value) && strings.TrimLeft(value, "0") != ""
}

func isDecimal(value string) bool {
	if value == "" {
		return false
	}
	for _, character := range value {
		if character < '0' || character > '9' {
			return false
		}
	}
	return true
}

func hasContentAddressedPath(segments []string) bool {
	for _, segment := range segments {
		tokens := strings.FieldsFunc(segment, func(character rune) bool {
			return character == '.' || character == '-' || character == '_' || character == '@'
		})
		for _, token := range tokens {
			minimumLength := 12
			if token == segment {
				minimumLength = 32
			}
			if len(token) >= minimumLength && len(token) <= 128 && isHex(token) && hasHexLetter(token) {
				return true
			}
		}
	}
	return false
}

func hasHexLetter(value string) bool {
	return strings.IndexFunc(value, func(character rune) bool {
		return (character >= 'a' && character <= 'f') || (character >= 'A' && character <= 'F')
	}) >= 0
}

func isHex(value string) bool {
	for _, character := range value {
		if !isHexCharacter(character) {
			return false
		}
	}
	return true
}

func isHexCharacter(character rune) bool {
	return (character >= '0' && character <= '9') ||
		(character >= 'a' && character <= 'f') ||
		(character >= 'A' && character <= 'F')
}
