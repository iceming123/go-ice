# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: gice deps android ios gice-cross swarm evm all test clean
.PHONY: gice-linux gice-linux-386 gice-linux-amd64 gice-linux-mips64 gice-linux-mips64le
.PHONY: gice-linux-arm gice-linux-arm-5 gice-linux-arm-6 gice-linux-arm-7 gice-linux-arm64
.PHONY: gice-darwin gice-darwin-386 gice-darwin-amd64
.PHONY: gice-windows gice-windows-386 gice-windows-amd64

GOBIN = $(shell pwd)/build/bin
GO ?= latest
DEPS = $(shell pwd)/internal/jsre/deps

gice:
	build/env.sh go run build/ci.go install ./cmd/gice
	@echo "Done building."
	@echo "Run \"$(GOBIN)/gice\" to launch gice."

genkey:
	$(GORUN) build/ci.go install ./cmd/genKey
	@echo "Done building."
	@echo "Run \"$(GOBIN)/genKey\" to launch genKey."

deps:
	cd $(DEPS) &&	go-bindata -nometadata -pkg deps -o bindata.go bignumber.js web3.js
	cd $(DEPS) &&	gofmt -w -s bindata.go
	@echo "Done generate deps."

swarm:
	build/env.sh go run build/ci.go install ./cmd/swarm
	@echo "Done building."
	@echo "Run \"$(GOBIN)/swarm\" to launch swarm."

all:
	build/env.sh go run build/ci.go install

# android:
#	build/env.sh go run build/ci.go aar --local
#	@echo "Done building."
#	@echo "Import \"$(GOBIN)/gice.aar\" to use the library."

# ios:
#	build/env.sh go run build/ci.go xcode --local
#	@echo "Done building."
#	@echo "Import \"$(GOBIN)/Gice.framework\" to use the library."

test: all
	build/env.sh go run build/ci.go test

lint: ## Run linters.
	build/env.sh go run build/ci.go lint

clean:
	rm -fr build/_workspace/pkg/ $(GOBIN)/*

# The devtools target installs tools required for 'go generate'.
# You need to put $GOBIN (or $GOPATH/bin) in your PATH to use 'go generate'.

devtools:
	env GOBIN= go get -u golang.org/x/tools/cmd/stringer
	env GOBIN= go get -u github.com/kevinburke/go-bindata/go-bindata
	env GOBIN= go get -u github.com/fjl/gencodec
	env GOBIN= go get -u github.com/golang/protobuf/protoc-gen-go
	env GOBIN= go install ./cmd/abigen
	@type "npm" 2> /dev/null || echo 'Please install node.js and npm'
	@type "solc" 2> /dev/null || echo 'Please install solc'
	@type "protoc" 2> /dev/null || echo 'Please install protoc'

# Cross Compilation Targets (xgo)

gice-cross: gice-linux gice-darwin gice-windows gice-android gice-ios
	@echo "Full cross compilation done:"
	@ls -ld $(GOBIN)/gice-*

gice-linux: gice-linux-386 gice-linux-amd64 gice-linux-arm gice-linux-mips64 gice-linux-mips64le
	@echo "Linux cross compilation done:"
	@ls -ld $(GOBIN)/gice-linux-*

gice-linux-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/386 -v ./cmd/gice
	@echo "Linux 386 cross compilation done:"
	@ls -ld $(GOBIN)/gice-linux-* | grep 386

gice-linux-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/amd64 -v ./cmd/gice
	@echo "Linux amd64 cross compilation done:"
	@ls -ld $(GOBIN)/gice-linux-* | grep amd64

gice-linux-arm: gice-linux-arm-5 gice-linux-arm-6 gice-linux-arm-7 gice-linux-arm64
	@echo "Linux ARM cross compilation done:"
	@ls -ld $(GOBIN)/gice-linux-* | grep arm

gice-linux-arm-5:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-5 -v ./cmd/gice
	@echo "Linux ARMv5 cross compilation done:"
	@ls -ld $(GOBIN)/gice-linux-* | grep arm-5

gice-linux-arm-6:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-6 -v ./cmd/gice
	@echo "Linux ARMv6 cross compilation done:"
	@ls -ld $(GOBIN)/gice-linux-* | grep arm-6

gice-linux-arm-7:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-7 -v ./cmd/gice
	@echo "Linux ARMv7 cross compilation done:"
	@ls -ld $(GOBIN)/gice-linux-* | grep arm-7

gice-linux-arm64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm64 -v ./cmd/gice
	@echo "Linux ARM64 cross compilation done:"
	@ls -ld $(GOBIN)/gice-linux-* | grep arm64

gice-linux-mips:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips --ldflags '-extldflags "-static"' -v ./cmd/gice
	@echo "Linux MIPS cross compilation done:"
	@ls -ld $(GOBIN)/gice-linux-* | grep mips

gice-linux-mipsle:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mipsle --ldflags '-extldflags "-static"' -v ./cmd/gice
	@echo "Linux MIPSle cross compilation done:"
	@ls -ld $(GOBIN)/gice-linux-* | grep mipsle

gice-linux-mips64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips64 --ldflags '-extldflags "-static"' -v ./cmd/gice
	@echo "Linux MIPS64 cross compilation done:"
	@ls -ld $(GOBIN)/gice-linux-* | grep mips64

gice-linux-mips64le:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips64le --ldflags '-extldflags "-static"' -v ./cmd/gice
	@echo "Linux MIPS64le cross compilation done:"
	@ls -ld $(GOBIN)/gice-linux-* | grep mips64le

gice-darwin: gice-darwin-386 gice-darwin-amd64
	@echo "Darwin cross compilation done:"
	@ls -ld $(GOBIN)/gice-darwin-*

gice-darwin-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=darwin/386 -v ./cmd/gice
	@echo "Darwin 386 cross compilation done:"
	@ls -ld $(GOBIN)/gice-darwin-* | grep 386

gice-darwin-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=darwin/amd64 -v ./cmd/gice
	@echo "Darwin amd64 cross compilation done:"
	@ls -ld $(GOBIN)/gice-darwin-* | grep amd64

gice-windows: gice-windows-386 gice-windows-amd64
	@echo "Windows cross compilation done:"
	@ls -ld $(GOBIN)/gice-windows-*

gice-windows-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=windows/386 -v ./cmd/gice
	@echo "Windows 386 cross compilation done:"
	@ls -ld $(GOBIN)/gice-windows-* | grep 386

gice-windows-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=windows/amd64 -v ./cmd/gice
	@echo "Windows amd64 cross compilation done:"
	@ls -ld $(GOBIN)/gice-windows-* | grep amd64
