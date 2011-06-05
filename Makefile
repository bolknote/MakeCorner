GOROOT=/Users/bolk/go
GDPATH=/usr/local/Cellar/gd/2.0.36RC1
JPEGTRAN=/usr/local/bin/jpegtran
GOARCH=amd64

include ${GOROOT}/src/Make.inc

TARG=corner

CGOFILES=\
	gd.go

CGO_CFLAGS=-I${GDPATH}/include
CGO_LDFLAGS+=-lgd

include ${GOROOT}/src/Make.pkg

all: install corner.go
	gofmt -r="\"JT\" -> \"$(JPEGTRAN)\"" jpegtran-template.go > jpegtran.go
	$(GC) ini.go
	$(GC) jpegtran.go
	$(GC) -I. corner.go
	$(LD) -L. -o corner corner.$O
