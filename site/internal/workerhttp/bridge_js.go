//go:build js && wasm

// Package workerhttp connects net/http handlers to the Worker's Fetch API.
package workerhttp

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"sync"
	"syscall/js"
)

// Register installs callbacks consumed by worker/runtime.js. Bodies and flushed
// chunks cross the boundary as Uint8Array, preserving binary bytes and UTF-8.
func Register(handler http.Handler) {
	var requests sync.Map
	js.Global().Set("goshtosoHandle", js.FuncOf(func(_ js.Value, args []js.Value) any {
		id, input, body, sink := args[0].String(), args[1], args[2], args[3]
		ctx, cancel := context.WithCancel(context.Background())
		requests.Store(id, cancel)
		go func() {
			defer requests.Delete(id)
			defer cancel()
			w := &responseWriter{header: make(http.Header), sink: sink, ctx: ctx}
			defer func() {
				if recovered := recover(); recovered != nil {
					sink.Call("fail", fmt.Sprint(recovered))
				}
			}()
			data := make([]byte, body.Length())
			js.CopyBytesToGo(data, body)
			r, err := http.NewRequestWithContext(ctx, input.Get("method").String(), input.Get("url").String(), bytes.NewReader(data))
			if err != nil {
				http.Error(w, "Invalid request", http.StatusBadRequest)
			} else {
				headers := input.Get("headers")
				for i := 0; i < headers.Length(); i++ {
					pair := headers.Index(i)
					r.Header.Add(pair.Index(0).String(), pair.Index(1).String())
				}
				if r.URL.Scheme == "https" {
					// Cookie consent and redirect logic can inspect the original HTTPS URL.
					r.Header.Set("X-Forwarded-Proto", "https")
				}
				handler.ServeHTTP(w, r)
			}
			w.flush(false)
			sink.Call("end")
		}()
		return nil
	}))
	js.Global().Set("goshtosoCancel", js.FuncOf(func(_ js.Value, args []js.Value) any {
		if value, ok := requests.Load(args[0].String()); ok {
			value.(context.CancelFunc)()
		}
		return nil
	}))
}

type responseWriter struct {
	header    http.Header
	committed http.Header
	status    int
	started   bool
	buffer    bytes.Buffer
	sink      js.Value
	ctx       context.Context
}

func (w *responseWriter) Header() http.Header { return w.header }

func (w *responseWriter) WriteHeader(status int) {
	if w.status != 0 {
		return
	}
	w.status = status
	w.committed = w.header.Clone()
}

func (w *responseWriter) Write(data []byte) (int, error) {
	if err := w.ctx.Err(); err != nil {
		return 0, err
	}
	if w.status == 0 {
		if w.header.Get("Content-Type") == "" {
			w.header.Set("Content-Type", http.DetectContentType(data))
		}
		w.WriteHeader(http.StatusOK)
	}
	return w.buffer.Write(data)
}

// Flush sends headers once and forwards pending bytes immediately. SSE handlers
// keep running until stream cancellation cancels the Go request context.
func (w *responseWriter) Flush() { w.flush(true) }

func (w *responseWriter) flush(streaming bool) {
	if w.ctx.Err() != nil {
		return
	}
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	if !w.started {
		headers := make([]any, 0, len(w.committed))
		for key, values := range w.committed {
			for _, value := range values {
				headers = append(headers, []any{key, value})
			}
		}
		w.sink.Call("start", w.status, js.ValueOf(headers), streaming)
		w.started = true
	}
	if w.buffer.Len() > 0 {
		chunk := js.Global().Get("Uint8Array").New(w.buffer.Len())
		js.CopyBytesToJS(chunk, w.buffer.Bytes())
		w.sink.Call("write", chunk)
		w.buffer.Reset()
	}
}
