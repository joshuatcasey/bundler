.PHONY: compile test retrieve metadata prepare

compile:
	@./compilation/compile.sh $(version) $(tarball_name)

test:
	@./test/test.sh $(tarball_name) $(version)

retrieve:
	@./retrieval/retrieve.sh

metadata:
	@./metadata/metadata.sh $(version)