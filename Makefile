GOROOT=/Users/bolk/go
GDPATH=/usr/local/Cellar/gd/2.0.36RC1
GOARCH=amd64

include ${GOROOT}/src/Make.inc

TARG=corner

CGOFILES=\
	gd.go

CGO_CFLAGS=-I${GDPATH}/include
CGO_LDFLAGS+=-L${GOPATH}/lib
CGO_LDFLAGS=-lgd

CLEANFILES+=ini corner jpegtran getncpu_$(GOOS)

include ${GOROOT}/src/Make.pkg

corner: install corner.go
	$(GC) ini.go
	$(GC) getncpu_$(GOOS).go
	$(GC) jpegtran.go
	$(GC) -I. corner.go
	$(LD) -L. -o $@ corner.$O
