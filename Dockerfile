FROM tinygo/tinygo:0.34.0 AS build

WORKDIR /work

COPY . .
RUN make build-debug

FROM traefik

COPY . /plugins-local/src/github.com/motoki317/traefik-rewritebody-wasm/
COPY --from=build /work/plugin.wasm /plugins-local/src/github.com/motoki317/traefik-rewritebody-wasm/
