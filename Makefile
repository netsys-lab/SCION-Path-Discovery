CC       = go build -v
BUILDDIR = ./bin
PRGS     = all



all: simple disjoint

.PHONY: disjoint
disjoint:
	$(CC) -o $(BUILDDIR)/$@ examples/disjoint/*.go

.PHONY: simple
simple:
	$(CC) -o $(BUILDDIR)/$@ examples/simple/*.go

clean:
	rm -f $(BUILDDIR)/*
