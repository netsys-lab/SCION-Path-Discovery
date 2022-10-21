CC       = go build -v
BUILDDIR = ./bin
PRGS     = all



all: simple disjoint mppingpong

.PHONY: disjoint
disjoint:
	$(CC) -o $(BUILDDIR)/$@ examples/disjoint/*.go

.PHONY: simple
simple:
	$(CC) -o $(BUILDDIR)/$@ examples/simple/*.go

.PHONY: mppingpong
mppingpong:
	$(CC) -o $(BUILDDIR)/$@ examples/mppingpong/*.go

clean:
	rm -f $(BUILDDIR)/*
