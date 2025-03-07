.PHONY: gen
gen:
	@echo "Generating code..."
	go generate ./...

.PHONY: clean-testdata
clean-testdata:
	git clean -xfd testdata

.PHONY: dev
dev:
	(cd testdata && go run ../cmd/preview/main.go web)