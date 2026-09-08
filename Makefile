.PHONY: build test race vet fmt web e2e smoke

build:            ## build all native binaries
	go build ./...

test:             ## run unit tests
	go test ./...

race:             ## run unit tests with the race detector
	go test -race ./...

vet:
	go vet ./...

fmt:
	gofmt -w cmd internal

web:              ## build the browser client into web/stk.wasm
	GOOS=js GOARCH=wasm go build -o web/stk.wasm ./cmd/stk-wasm

e2e:              ## run the containerized end-to-end test
	docker compose -f deploy/compose/compose.yml build
	docker compose -f deploy/compose/compose.yml run --rm e2e
	docker compose -f deploy/compose/compose.yml down -v

smoke:            ## run the no-Docker local end-to-end smoke test
	./smoke.sh
