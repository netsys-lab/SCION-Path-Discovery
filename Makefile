CC       = go build -v
BUILDDIR = ./bin
PRGS     = simple



.PHONY: nico-simple
nico-simple:
	mkdir -p $(BUILDDIR)
	$(CC) -o $(BUILDDIR)/simple examples/simple/*.go
	cp -f $(BUILDDIR)/simple /home/nicolas/Documents/BitTorrent/VM_e9e
	cp -f $(BUILDDIR)/simple /home/nicolas/Documents/BitTorrent/VM_ea6

.PHONY: simple
simple: prepare
	$(CC) -o $(BUILDDIR)/$@ examples/simple/*.go


.PHONY: karola-simple
karola-simple:
	mkdir -p $(BUILDDIR)
	$(CC) -o $(BUILDDIR)/$@ examples/simple/*.go
	cp -f $(BUILDDIR)/simple /mnt/k/EU
	cp -f $(BUILDDIR)/simple /mnt/k/NA

clean:
	rm -f $(BUILDDIR)/*
