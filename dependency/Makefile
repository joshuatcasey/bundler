.PHONY: test retrieve

id:
	@echo bundler

test:
	@./test/test.sh $(tarball_name) $(version)

retrieve:
	@cd retrieve; \
	go build -o retrieve; \
	./retrieve --buildpack_toml_path=$(buildpack_toml_path) \
			--metadata_file=$(metadata_file) \
			--targets_file=$(targets_file); \
	rm retrieve

