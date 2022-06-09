.PHONY: compile test retrieve metadata prepare

compile:
	@./compilation/compile.sh $(version)

test:
	@./test/test.sh $(tarball_path) $(version) $(os)

retrieve:
	@./retrieval/retrieve.sh

metadata:
	@./metadata/metadata.sh $(version)

prepare:
	@./prepare/prepare.sh $(os)
