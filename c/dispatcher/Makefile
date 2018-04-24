.PHONY: all clean install uninstall

GOLANG = go
CFLAGS ?= -Wall -Werror -g -O2
LDFLAGS ?= -L/usr/local/lib -lzlog -L/usr/local/lib -ltcpmw -L/usr/local/lib -llwip -L/usr/local/lib -lfilter -L/usr/local/lib -lscion -lpthread

LIB_DIR = $(realpath ../lib)
LWIPDIR = $(realpath ../../sub/lwip/src/include)
LWIPPORTDIR = $(realpath ../../sub/lwip-contrib/ports/unix)
INC = -I$(LIB_DIR) -I$(LWIPDIR) -I$(LWIPDIR)/scion -I$(LWIPDIR)/ipv4
INC += -I$(LWIPPORTDIR)/include -I$(LWIPPORTDIR)/proj/scion
INC += -I/usr/local/include

TARGETS = dispatcher
INSTALL = $(realpath ../../bin/dispatcher)
PREFIX = ..
all: $(TARGETS)

clean:
	rm -f *.d $(TARGETS)

# Compile .c files, while also automatically generating dependencies so they
# can be recompiled if any of the included header files changed.
-include *.d
dispatcher: dispatcher.go
	CGO_CFLAGS="$(INC)" CGO_LDFLAGS="$(LDFLAGS)" $(GOLANG) build $<

install: $(INSTALL)

$(INSTALL): dispatcher
	cp -f dispatcher $@

uninstall:
	rm -f $(INSTALL)
