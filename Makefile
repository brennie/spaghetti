PKGS=. ./pqueue ./set ./solver ./solver/hpga ./tt

.PHONY: build clean format

build:
	go build

clean:
	@rm -f spaghetti spaghetti.exe

format:
	go fmt $(PKGS)
