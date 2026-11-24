GO ?= go

.PHONY: build run test vet demo demo-copy run-interpreter clean

build:
	mkdir -p bin
	$(GO) build -o bin/assembler ./src
	$(GO) build -o bin/interpreter ./src/interpreter

run:
	$(GO) run ./src $(ARGS)

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

demo:
	$(GO) run ./src -input examples/specification.asm -output program.bin -test

run-interpreter:
	$(GO) run ./src/interpreter $(ARGS)

demo-copy: build
	mkdir -p build
	./bin/assembler -input examples/copy_array.asm -output build/copy.bin
	./bin/interpreter -input build/copy.bin -output build/copy.csv -start 2000 -end 2004
	diff -u tests/copy_array.expected.csv build/copy.csv
	cat build/copy.csv

clean:
	rm -rf bin build
