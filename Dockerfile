# Working version notes:
#
# 0.30.0 -> does not support go 1.23 (supports 1.21 at max), while latest golang.org/x/text requires 1.23
#
# 0.33.0 -> working fine locally with go 1.23
#
# 0.34.0 -> "out of memory reading body" error for some reason, when we use transform.NewReader
# github.com/traefik/traefik/v3/pkg/middlewares/recovery/recovery.go:49 > Recovered from panic in HTTP handler [172.20.0.1:59028 - /]: out of memory reading body (recovered by wazero)
#
# >=0.35.0 -> http-wasm does not support newer versions
# https://github.com/traefik/traefik/issues/11916
FROM tinygo/tinygo:0.33.0 AS build

WORKDIR /work

COPY . .
RUN make build-debug

FROM traefik

COPY . /plugins-local/src/github.com/motoki317/traefik-rewritebody-wasm/
COPY --from=build /work/plugin.wasm /plugins-local/src/github.com/motoki317/traefik-rewritebody-wasm/
