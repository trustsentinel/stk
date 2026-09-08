# Builds the stk MVP binaries (hub, agent, client, keygen) and packages them in a
# small runtime image with the tools the e2e test needs (curl, nc). Build context
# is the repo root so it can copy go.mod + the source. _legacy/ and crypto/ are
# not copied, so only the MVP module is built.
FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN set -eux; \
    for c in stk-hub stk-agent stk-client stk-keygen; do \
      CGO_ENABLED=0 go build -o "/out/$c" "./cmd/$c"; \
    done

FROM alpine:3.20
RUN apk add --no-cache curl netcat-openbsd
COPY --from=build /out/ /usr/local/bin/
