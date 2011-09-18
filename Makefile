GOROOT=/usr/local/Cellar/go/r59
GDPATH=/usr/local/Cellar/gd/2.0.36RC1
JPEGTRAN=/usr/local/bin/jpegtran

include ${GOROOT}/src/Make.inc

TARG=corner

CGOFILES=\
	gd.go

CGO_CFLAGS=-I${GDPATH}/include
CGO_LDFLAGS+=-lgd

CLEANFILES+=jpegtran.go

include ${GOROOT}/src/Make.pkg

ifeq ($(GOOS),windows)
	EXT=.exe
else
	EXT=
endif

all: corner.go ini.$O jpegtran.$O corner.$O
	$(LD) -L. -o corner$(EXT) corner.$O

%.$(O): %.go
	$(GC) -I. $^

jpegtran.go:
	@echo "package jpegtran; const Jpegtran=\`$(JPEGTRAN)\`" > jpegtran.go
