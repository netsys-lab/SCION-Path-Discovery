CC       = go build
BUILDDIR = ./bin



.PHONY: nico-simple
nico-simple:
	mkdir -p $(BUILDDIR)
	$(CC) -o $(BUILDDIR)/simple examples/simple/*.go
	cp -f $(BUILDDIR)/simple /mnt/k/EU
	cp -f $(BUILDDIR)/simple /mnt/k/NA



.PHONY: karola-simple
karola-simple:
	mkdir -p $(BUILDDIR)
	$(CC) -o $(BUILDDIR)/$@ examples/simple/*.go
	cp -f $(BUILDDIR)/simple /mnt/k/Dokumente/EU
	cp -f $(BUILDDIR)/simple /mnt/k/Dokumente/NA

clean:
	rm -f $(BUILDDIR)/*