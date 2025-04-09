GO_SRC_FILES := $(shell find . $(FIND_EXCLUSIONS) -type f -name '*.go' -not -name '*_test.go')

.PHONY: gen
gen:
	@echo "Generating code..."
	go generate ./...

.PHONY: clean-testdata
clean-testdata:
	git clean -xfd testdata


.PHONY: build-wasm
build-wasm: build/preview.wasm
	mkdir -p ./build

build/preview.wasm: $(GO_SRC_FILES)
	GOOS=js GOARCH=wasm go build -o build/preview.wasm