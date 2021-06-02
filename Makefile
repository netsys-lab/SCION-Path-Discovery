CC       = go build
BUILDDIR = ./bin
PRGS     = simple



all: $(PRGS)

.PHONY: simple
simple: prepare
	$(CC) -o $(BUILDDIR)/$@ examples/simple/*.go



.PHONY: prepare
prepare:
	mkdir -p $(BUILDDIR)



clean:
	rm -f $(BUILDDIR)/*
