GO ?= go

.PHONY: build run test vet demo clean

build:
	mkdir -p bin
	$(GO) build -o bin/assembler ./src

run:
	$(GO) run ./src $(ARGS)

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

demo:
	$(GO) run ./src -input examples/specification.asm -output program.bin -test

clean:
	rm -rf bin
