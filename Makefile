GOCMD   = go
GOBUILD = $(GOCMD) build
GOFMT   = $(GOCMD) fmt
GOCLEAN = $(GOCLEAN)
GOPATH  = ${PWD}:${GOPATHi}
export $(GOPATH)

SRCDIR   = ${PWD}/src
BINDIR   = ${PWD}/bin
MANDIR   = ${PWD}/man
COMPDIR  = ${PWD}/completion
PACKAGE  = kcp
SRCFILE  = $(SRCDIR)/$(PACKAGE)/$(PACKAGE).go
BINFILE  = $(BINDIR)/$(PACKAGE)
MANFILE  = $(MANDIR)/$(PACKAGE).1
BASHFILE = $(COMPDIR)/$(PACKAGE).bash

DESTDIR    = .
INSTALLDIR = $(DESTDIR)/usr

DESTMAN  = $(INSTALLDIR)/share/man/man1
DESTBIN  = $(INSTALLDIR)/bin
DESTBASH = $(DESTDIR)/etc/bash_completion.d

.PHONY: default build install clean

default: build

build:
	$(GOBUILD) -v -o $(BINFILE) $(SRCFILE)
	mkdir -p $(MANDIR)
	$(BINFILE) --create-manpage > $(MANFILE)

clean:
	rm -rf $(BINDIR) $(MANDIR)

install:
	mkdir -p $(DESTBIN) $(DESTMAN) $(DESTBASH)
	cp $(BINFILE) $(DESTBIN)
	cp $(MANFILE) $(DESTMAN)
	cp $(BASHFILE) $(DESTBASH)/$(PACKAGE)

