.PHONY: test retrieve metadata

test:
	@./test/test.sh $(tarball_name) $(version)

retrieve:
	@./retrieval/retrieve.sh $(buildpackTomlPath) $(output)

metadata:
	@./metadata/metadata.sh $(version) $(id) $(name) $(output)
