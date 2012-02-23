GOROOT=/usr/local/Cellar/go/r60.3
JPEGTRAN=/usr/local/bin/jpegtran
GOGDPATH=github.com/bolknote/go-gd

include ${GOROOT}/src/Make.inc
TARG=corner
CLEANFILES+=jpegtran.go gd.go
CGOFILES=gd.go
CGO_LDFLAGS=-lgd

include ${GOROOT}/src/Make.pkg

ifeq ($(GOOS),windows)
	EXT=.exe
else
	EXT=
endif

all: corner.go ini.$O jpegtran.$O corner.$O gd.go
	$(LD) -L. -o corner$(EXT) corner.$O

%.$(O): %.go
	$(GC) -I. $^

jpegtran.go:
	@echo "package jpegtran; const Jpegtran=\`$(JPEGTRAN)\`" > jpegtran.go

gd.go:
	CGO_LDFLAGS=-lgd GOPATH=${GOROOT} goinstall ${GOGDPATH}
	cp -f ${GOROOT}/src/${GOGDPATH}/gd.go .
