CC       = go build -v
BUILDDIR = ./bin
PRGS     = all



all: simple mppingpong

.PHONY: mppingpong
mppingpong:
	$(CC) -o $(BUILDDIR)/$@ examples/mppingpong/*.go

.PHONY: simple
simple:
	$(CC) -o $(BUILDDIR)/$@ examples/simple/*.go

clean:
	rm -f $(BUILDDIR)/*
