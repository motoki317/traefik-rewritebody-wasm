# Traefik Rewrite Body WASM Plugin

A Traefik middleware plugin that rewrites HTTP response bodies using WebAssembly.

## Features

- Stream-based body rewriting for efficient memory usage
- Multiple string replacements in a single pass
- WASM-based implementation for security and portability

## Installation

### From Plugin Catalog

```yaml
# Static configuration
experimental:
  plugins:
    rewritebody:
      moduleName: github.com/motoki317/traefik-rewritebody-wasm
      version: v0.1.0
```

### Local Development

```yaml
# Static configuration
experimental:
  localPlugins:
    rewritebody:
      moduleName: github.com/motoki317/traefik-rewritebody-wasm
```

## Configuration

```yaml
# Dynamic configuration
http:
  middlewares:
    my-rewriter:
      plugin:
        rewritebody:
          rewrites:
            - from: "old-domain.com"
              to: "new-domain.com"
            - from: "http://"
              to: "https://"
```

## Options

- `rewrites`: Array of rewrite rules
  - `from`: String to search for
  - `to`: String to replace with

## Example

```yaml
http:
  routers:
    my-router:
      rule: Host(`example.com`)
      middlewares:
        - my-rewriter
      service: my-service

  middlewares:
    my-rewriter:
      plugin:
        rewritebody:
          rewrites:
            - from: "localhost:8080"
              to: "example.com"
            - from: "http://api.internal"
              to: "https://api.example.com"

  services:
    my-service:
      loadBalancer:
        servers:
          - url: http://localhost:8080
```

## Building

```bash
make build
```

## License

MIT
