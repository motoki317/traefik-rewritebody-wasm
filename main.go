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
	features := handler.Host.EnableFeatures(api.FeatureBufferResponse)
	handler.Host.Log(api.LogLevelInfo, "[DEBUG] Features enabled: "+features.String())

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
	handler.Host.Log(api.LogLevelInfo, "Loaded plugin")
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

func (p *Plugin) chain(w io.Writer) io.Writer {
	ret := w
	for i := len(p.replacers) - 1; i >= 0; i-- {
		ret = transform.NewWriter(ret, p.replacers[i])
	}
	return ret
}

func (p *Plugin) handleResponse(_ uint32, req api.Request, resp api.Response, isError bool) {
	// Only process successful responses
	if isError {
		return
	}

	handler.Host.Log(api.LogLevelInfo, "Processing response for url="+req.GetURI())

	// Create wrappers for WASI Body interface
	writer := NewBodyWriter(resp.Body())

	// Create the transformer chain with our reader
	transformer := p.chain(writer)

	// Copy the transformed content to the writer
	_, err := resp.Body().WriteTo(transformer)
	if err != nil {
		handler.Host.Log(api.LogLevelError, fmt.Sprintf("Could not write response %v", err))
	}
}
