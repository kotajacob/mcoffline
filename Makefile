# mcoffline
# See LICENSE for copyright and license details.
.POSIX:
.SUFFIXES:

VERSION = 1.0.0
GO = go
RM = rm
INSTALL = install
SCDOC = scdoc
GOFLAGS =
PREFIX = /usr/local
BINDIR = bin
MANDIR = share/man

all: mcoffline doc/mcoffline.1

mcoffline:
	$(GO) build -ldflags "-X main.Version=$(VERSION)" $(GOFLAGS)

doc/mcoffline.1: doc/mcoffline.1.scd
	$(SCDOC) < doc/mcoffline.1.scd | sed "s/VERSION/$(VERSION)/g" > doc/mcoffline.1

clean:
	$(RM) -f mcoffline doc/mcoffline.1

install:
	$(INSTALL) -dp \
		$(DESTDIR)$(PREFIX)/$(BINDIR)/ \
		$(DESTDIR)$(PREFIX)/$(MANDIR)/man1/
	$(INSTALL) -pm 0755 mcoffline -t $(DESTDIR)$(PREFIX)/$(BINDIR)/
	$(INSTALL) -pm 0644 doc/mcoffline.1 -t $(DESTDIR)$(PREFIX)/$(MANDIR)/man1/

uninstall:
	$(RM) -f \
		$(DESTDIR)$(PREFIX)/$(BINDIR)/mcoffline \
		$(DESTDIR)$(PREFIX)/$(MANDIR)/man1/mcoffline.1

.PHONY: all mcofflinmcoffline install uninstall
