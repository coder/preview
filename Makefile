.PHONY: gen
gen:
	@echo "Generating code..."
	go generate ./...

.PHONY: clean-testdata
clean-testdata:
	git clean -xfd testdata