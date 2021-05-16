# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

.PYHONY: build
# Build plugin binary
build: fmt vet
	CGO_ENABLED=0 GO111MODULE=on go build -o bin/adopt ./main.go

# Run go fmt against code
.PYHONY: fmt
fmt:
	go fmt ./...

# Run go vet against code
.PYHONY: vet
vet:
	go vet ./...
