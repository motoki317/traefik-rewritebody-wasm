package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

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
	handler.Host.Log(api.LogLevelInfo, fmt.Sprintf("[traefik-rewritebody-wasm] Loaded plugin with %d rewrite(s), allowed content types: %v", len(config.Rewrites), mw.allowedContentTypes))
}

// Config is the plugin configuration.
type Config struct {
	AllowedContentTypes []string  `json:"allowedContentTypes"`
	Rewrites            []Rewrite `json:"rewrites"`
}

type Rewrite struct {
	From string `json:"from"`
	To   string `json:"to"`
}

func (c *Config) createTransformer() transform.Transformer {
	replacers := make([]transform.Transformer, len(c.Rewrites))
	for i, rewrite := range c.Rewrites {
		replacers[i] = replace.String(rewrite.From, rewrite.To)
	}
	if len(replacers) == 0 {
		return transform.Nop
	}
	if len(replacers) == 1 {
		return replacers[0]
	}
	return transform.Chain(replacers...)
}

// Plugin is a plugin instance.
type Plugin struct {
	allowedContentTypes []string
	replacer            transform.Transformer
}

// New creates a new plugin instance.
func New(config Config) (*Plugin, error) {
	return &Plugin{
		allowedContentTypes: config.AllowedContentTypes,
		replacer:            config.createTransformer(),
	}, nil
}

func (p *Plugin) isAllowedType(contentType string) bool {
	if len(p.allowedContentTypes) == 0 {
		return true
	}
	for _, allowedType := range p.allowedContentTypes {
		if strings.Contains(contentType, allowedType) {
			return true
		}
	}
	return false
}

func (p *Plugin) handleResponse(_ uint32, req api.Request, resp api.Response, isError bool) {
	// Only process successful responses
	if isError {
		return
	}
	// Only process configured content-types
	if contentType, ok := resp.Headers().Get("Content-Type"); ok && !p.isAllowedType(contentType) {
		return
	}

	handler.Host.Log(api.LogLevelDebug, "Processing rewrite for uri: "+req.GetURI())

	// Create wrappers for WASI Body interface
	reader := NewBodyReader(resp.Body())
	writer := NewBodyWriter(resp.Body())

	// Create the transformer chain with our reader
	transformed := transform.NewReader(reader, p.replacer)

	// Copy the transformed content to the writer
	buf := make([]byte, 2*1024) // 2KiB buffer size is excepted from resp.Body().WriteTo() implementation.
	n, err := io.CopyBuffer(writer, transformed, buf)
	if err != nil {
		handler.Host.Log(api.LogLevelError, fmt.Sprintf("Could not write response %v", err))
	}
	resp.Headers().Remove("Content-Length")
	resp.Headers().Add("Content-Length", fmt.Sprintf("%d", n))
}
