CC       = go build
BUILDDIR = ./bin



.PHONY: nico-simple
nico-simple:
	mkdir -p $(BUILDDIR)
	$(CC) -o $(BUILDDIR)/simple examples/simple/*.go
	cp -f $(BUILDDIR)/simple /mnt/k/Dokumente/Teamprojekt/BitTorrent/1
	cp -f $(BUILDDIR)/simple /mnt/k/Dokumente/Teamprojekt/BitTorrent/2



.PHONY: karola-simple
karola-simple:
	mkdir -p $(BUILDDIR)
	$(CC) -o $(BUILDDIR)/$@ examples/simple/*.go
	cp -f $(BUILDDIR)/simple /mnt/k/Dokumente/Teamprojekt/BitTorrent/1
	cp -f $(BUILDDIR)/simple /mnt/k/Dokumente/Teamprojekt/BitTorrent/2


clean:
	rm -f $(BUILDDIR)/*