GOROOT=/Users/bolk/go
GDPATH=/usr/local/Cellar/gd/2.0.36RC1
GOARCH=amd64

include ${GOROOT}/src/Make.inc

TARG=corner

CGOFILES=\
	gd.go

CGO_CFLAGS=-I${GDPATH}/include
CGO_LDFLAGS+=-lgd

include ${GOROOT}/src/Make.pkg

corner: install corner.go
	$(GC) jpegtran-convert.go
	$(LD) -o jpegtran-convert jpegtran-convert.$O
	./jpegtran-convert && unlink jpegtran-convert
	$(GC) ini.go
	$(GC) -o getncpu.$O getncpu_$(GOOS).go
	$(GC) jpegtran.go
	$(GC) -I. corner.go
	$(LD) -L. -o $@ corner.$O
