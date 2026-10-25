//go:build js && wasm

// The Go server runs inside Cloudflare Workers, never in the browser.
package main

import (
	"encoding/json"
	"syscall/js"

	"github.com/araihu/goshtoso/site/internal/server"
	"github.com/araihu/goshtoso/site/internal/workerhttp"
)

func main() {
	workerhttp.Register(server.NewWorker())
	js.Global().Set("goshtosoChat", js.FuncOf(func(_ js.Value, args []js.Value) any {
		var input server.WorkerChatRequest
		if err := json.Unmarshal([]byte(args[0].String()), &input); err != nil {
			return "null"
		}
		output, err := json.Marshal(server.WorkerChat(input))
		if err != nil {
			return "null"
		}
		return string(output)
	}))
	select {}
}
