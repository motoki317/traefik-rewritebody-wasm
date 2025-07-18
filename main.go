package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/http-wasm/http-wasm-guest-tinygo/handler"
	"github.com/http-wasm/http-wasm-guest-tinygo/handler/api"
	"github.com/icholy/replace"
	"golang.org/x/text/transform"
)

func main() {
	var config Config
	err := json.Unmarshal(handler.Host.GetConfig(), &config)
	if err != nil {
		handler.Host.Log(api.LogLevelError, fmt.Sprintf("Could not load config %v", err))
		os.Exit(1)
	}

	mw, err := New(config)
	if err != nil {
		handler.Host.Log(api.LogLevelError, fmt.Sprintf("Could not load config %v", err))
		os.Exit(1)
	}
	handler.HandleRequestFn = mw.handleRequest
	handler.HandleResponseFn = mw.handleResponse
}

// Config is the plugin configuration.
type Config struct {
	Rewrites []Rewrite `json:"rewrites"`
}

type Rewrite struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// Plugin is a plugin instance.
type Plugin struct {
	replacers []transform.Transformer
}

// New creates a new plugin instance.
func New(config Config) (*Plugin, error) {
	replacers := make([]transform.Transformer, 0, len(config.Rewrites))
	for _, rewrite := range config.Rewrites {
		replacers = append(replacers, replace.String(rewrite.From, rewrite.To))
	}
	return &Plugin{
		replacers: replacers,
	}, nil
}

func (p *Plugin) handleRequest(_ api.Request, _ api.Response) (next bool, reqCtx uint32) {
	// Pass through to the next handler
	return true, 0
}

func (p *Plugin) handleResponse(_ uint32, _ api.Request, resp api.Response, isError bool) {
	// Only process successful responses
	if isError {
		return
	}

	// Create wrappers for WASI Body interface
	reader := NewBodyReader(resp.Body())
	writer := NewBodyWriter(resp.Body())

	// Create the transformer chain with our reader
	transformer := replace.Chain(reader, p.replacers...)

	// Copy the transformed content to the writer
	_, _ = io.Copy(writer, transformer)
}
