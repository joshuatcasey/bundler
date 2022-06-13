.PHONY: compile test retrieve metadata

compile:
	@./compilation/compile.sh $(version) $(tarball_name) $(os)

test:
	@./test/test.sh $(tarball_name) $(version)

retrieve:
	@./retrieval/retrieve.sh

metadata:
	@./metadata/metadata.sh $(version)
