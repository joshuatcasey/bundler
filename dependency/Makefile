.PHONY: test retrieve

test:
	@./test/test.sh $(tarball_name) $(version)

retrieve:
	@pushd retrieve > /dev/null; \
	go build -o retrieve; \
	./retrieve --buildpack_toml_path=$(buildpack_toml_path) \
			--output_file=$(output_file) \
			--id=$(id) \
			--name=$(name); \
	rm retrieve; \
	popd > /dev/null

