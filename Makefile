CC       = go build
BUILDDIR = ./bin



.PHONY: nico-simple
nico-simple:
	mkdir -p $(BUILDDIR)
	$(CC) -o $(BUILDDIR)/simple examples/simple/*.go
	cp -f $(BUILDDIR)/simple /mnt/k/148
	cp -f $(BUILDDIR)/simple /mnt/k/151



.PHONY: karola-simple
karola-simple:
	mkdir -p $(BUILDDIR)
	$(CC) -o $(BUILDDIR)/$@ examples/simple/*.go
	cp -f $(BUILDDIR)/simple /mnt/k/Dokumente/148
	cp -f $(BUILDDIR)/simple /mnt/k/Dokumente/151

clean:
	rm -f $(BUILDDIR)/*