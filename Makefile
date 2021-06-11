CC       = go build
BUILDDIR = ./bin



.PHONY: nico-simple
nico-simple:
	mkdir -p $(BUILDDIR)
	$(CC) -o $(BUILDDIR)/simple examples/simple/*.go
	cp -f $(BUILDDIR)/simple /home/nicolas/Documents/BitTorrent/VM_e9e
	cp -f $(BUILDDIR)/simple /home/nicolas/Documents/BitTorrent/VM_ea6



.PHONY: karola-simple
karola-simple:
	mkdir -p $(BUILDDIR)
	$(CC) -o $(BUILDDIR)/$@ examples/simple/*.go
	#cp -f $(BUILDDIR)/$@ /home/nicolas/Documents/BitTorrent/VM_e9e



clean:
	rm -f $(BUILDDIR)/*
