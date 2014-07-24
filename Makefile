GOCMD   = go
GOBUILD = $(GOCMD) build
GOFMT   = $(GOCMD) fmt
GOCLEAN = $(GOCLEAN)
GOPATH  = ${PWD}:${GOPATHi}
export $(GOPATH)

SRCDIR   = src
BINDIR   = bin
MANDIR   = man
PACKAGE  = kcp
SRCFILE  = $(SRCDIR)/$(PACKAGE)/$(PACKAGE).go
BINFILE  = $(BINDIR)/$(PACKAGE)
MANFILE  = $(MANDIR)/$(PACKAGE).1

DESTDIR    = .
INSTALLDIR = $(DESTDIR)/usr

DESTMAN = $(INSTALLDIR)/share/man/man1
DESTBIN = $(INSTALLDIR)/bin

.PHONY: default format build man install clean

default: build man

build: format
	$(GOBUILD) -v -o $(BINFILE) $(SRCFILE)

format:
	$(GOFMT) $(SRCDIR)/...

clean:
	rm -rf $(BINDIR) $(MANDIR)

man:
	mkdir -p $(MANDIR)
	$(BINFILE) --create-manpage > $(MANFILE)

install:
	mkdir -p $(DESTBIN) $(DESTMAN)
	cp $(BINFILE) $(DESTBIN)
	cp $(MANFILE) $(DESTMAN)
