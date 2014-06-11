PKGS=. ./pqueue ./set ./solver ./solver/hpga ./tt

.PHONY: build test clean format

build:
	go build

test:
	go test ./set

clean:
	@rm -f spaghetti spaghetti.exe

format:
	go fmt $(PKGS)
