package assets

import (
	"io"
	"net/http"
	"strings"
)

type muambaRef struct {
	resource string
	download string
}

var muambaHTTPFiles = muambaHTTPIndex()

func muambaHTTPIndex() map[string]muambaRef {
	index := make(map[string]muambaRef)
	for _, resource := range MuambaResources() {
		for _, download := range resource.Downloads {
			path, ok := strings.CutPrefix(download.Path, "assets/")
			if ok {
				index[path] = muambaRef{resource: resource.Name, download: download.Name}
			}
		}
	}
	return index
}

func serveMuambaFile(writer http.ResponseWriter, request *http.Request, ref muambaRef) {
	file, err := MuambaOpen(ref.resource, ref.download)
	if err != nil {
		http.NotFound(writer, request)
		return
	}
	defer func() { _ = file.Close() }()

	seeker, ok := file.(io.ReadSeeker)
	if !ok {
		http.Error(writer, "embedded asset is not seekable", http.StatusInternalServerError)
		return
	}
	info, err := file.Stat()
	if err != nil {
		http.Error(writer, "embedded asset metadata unavailable", http.StatusInternalServerError)
		return
	}
	http.ServeContent(writer, request, info.Name(), info.ModTime(), seeker)
}
