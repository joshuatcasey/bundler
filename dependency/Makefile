.PHONY: compile test retrieve metadata

compile:
	@./compilation/compile.sh $(version)

test:
	@./test/test.sh $(tarball_path) $(version)

retrieve:
	@./retrieval/retrieve

annotate:
	@./metadata/annotate.sh
