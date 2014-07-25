GOCMD   = go
GOBUILD = $(GOCMD) build
GOFMT   = $(GOCMD) fmt
GOCLEAN = $(GOCLEAN)
GOPATH  = ${PWD}:${GOPATHi}
export $(GOPATH)

SRCDIR   = ${PWD}/src
BINDIR   = ${PWD}/bin
MANDIR   = ${PWD}/man
PACKAGE  = kcp
SRCFILE  = $(SRCDIR)/$(PACKAGE)/$(PACKAGE).go
BINFILE  = $(BINDIR)/$(PACKAGE)
MANFILE  = $(MANDIR)/$(PACKAGE).1

DESTDIR    = .
INSTALLDIR = $(DESTDIR)/usr

DESTMAN = $(INSTALLDIR)/share/man/man1
DESTBIN = $(INSTALLDIR)/bin

.PHONY: default build install clean

default: build

build:
	$(GOFMT) $(SRCDIR)/...
	$(GOBUILD) -v -o $(BINFILE) $(SRCFILE)
	mkdir -p $(MANDIR)
	$(BINFILE) --create-manpage > $(MANFILE)

clean:
	rm -rf $(BINDIR) $(MANDIR)

install:
	mkdir -p $(DESTBIN) $(DESTMAN)
	cp $(BINFILE) $(DESTBIN)
	cp $(MANFILE) $(DESTMAN)
