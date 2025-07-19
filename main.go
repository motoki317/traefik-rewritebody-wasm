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
	handler.Host.EnableFeatures(api.FeatureBufferResponse)

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
	handler.HandleResponseFn = mw.handleResponse
	handler.Host.Log(api.LogLevelInfo, fmt.Sprintf("Loaded plugin with %d rewrite(s)", len(config.Rewrites)))
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
	replacer transform.Transformer
}

// New creates a new plugin instance.
func New(config Config) (*Plugin, error) {
	replacers := make([]transform.Transformer, len(config.Rewrites))
	for i, rewrite := range config.Rewrites {
		replacers[i] = replace.String(rewrite.From, rewrite.To)
	}
	if len(replacers) == 0 {
		return &Plugin{replacer: transform.Nop}, nil
	}
	if len(replacers) == 1 {
		return &Plugin{replacer: replacers[0]}, nil
	}
	return &Plugin{
		replacer: transform.Chain(replacers...),
	}, nil
}

func (p *Plugin) handleResponse(_ uint32, req api.Request, resp api.Response, isError bool) {
	// Only process successful responses
	if isError {
		return
	}

	handler.Host.Log(api.LogLevelInfo, "Processing response for url="+req.GetURI())

	// Create wrappers for WASI Body interface
	reader := NewBodyReader(resp.Body())
	writer := NewBodyWriter(resp.Body())

	// Create the transformer chain with our reader
	transformed := transform.NewReader(reader, p.replacer)

	// Copy the transformed content to the writer
	buf := make([]byte, 2*1024)
	n, err := io.CopyBuffer(writer, transformed, buf)
	if err != nil {
		handler.Host.Log(api.LogLevelError, fmt.Sprintf("Could not write response %v", err))
	}
	resp.Headers().Remove("Content-Length")
	resp.Headers().Add("Content-Length", fmt.Sprintf("%d", n))
}
