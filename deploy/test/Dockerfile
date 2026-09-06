# Containerized tests for the stk crypto module.
#
# This is the part of stk that builds and passes TODAY. The full
# broker/agent/web stack does not build yet (2019 GOPATH prototype +
# defunct gopkg.in/noisesocket.v0) — see ../../docs/PORTING.md and the
# target end-to-end scenario in ../e2e.blueprint.yml.
#
# The crypto module is stdlib-only, so this build needs no network for deps.
FROM golang:1.21-alpine

WORKDIR /src
COPY crypto/ ./crypto/
WORKDIR /src/crypto

# Build gate: if the module stops compiling, the image build fails here.
RUN go vet ./...

# Tests run at container start so `docker compose run` streams the output
# and the exit code reflects pass/fail.
CMD ["go", "test", "-v", "./..."]
