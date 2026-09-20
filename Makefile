GO ?= go

.PHONY: all test race fuzz cover fmt vet lint build clean

# The default target: format, vet, test.
all: fmt vet test

build:
	$(GO) build ./...

test:
	$(GO) test ./...

# The race detector is the point of TestConcurrentUse.
race:
	$(GO) test -race ./...

# Property-based fuzzing of the round-trip guarantee. Override the budget with
# make fuzz FUZZTIME=2m
FUZZTIME ?= 30s
fuzz:
	$(GO) test -run '^$$' -fuzz FuzzRoundTrip -fuzztime $(FUZZTIME) .

cover:
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out

fmt:
	gofmt -l -w .

vet:
	$(GO) vet ./...

lint:
	golangci-lint run

clean:
	rm -f coverage.out