# Constants
MAIN_NAME = d-bf
EXT_LINUX = .bin
EXT_WINDOWS = .exe
EXT_MAC = .app

CFLAGS = -std=gnu99 -O2 -lm -lcurl ./src/lib/cJSON/cJSON.c ./src/lib/base64/base64.c

CC_LINUX32        = gcc
CC_LINUX64        = gcc
CC_WINDOWS32      = /usr/bin/i686-w64-mingw32-gcc
CC_WINDOWS64      = /usr/bin/x86_64-w64-mingw32-gcc
CC_OSX32          = /usr/bin/i686-apple-darwin10-gcc
CC_OSX64          = /usr/bin/i686-apple-darwin10-gcc

CFLAGS_LINUX32    = $(CFLAGS) -m32 -DLINUX		-I./cross_compile/linux/32/include/
CFLAGS_LINUX64    = $(CFLAGS) -m64 -DLINUX		-I./cross_compile/linux/64/include/
CFLAGS_WINDOWS32  = $(CFLAGS) -m32 -DWINDOWS	-I./cross_compile/windows/32/include/	-L./cross_compile/windows/32/lib/
CFLAGS_WINDOWS64  = $(CFLAGS) -m64 -DWINDOWS	-I./cross_compile/windows/64/include/	-L./cross_compile/windows/64/lib/
CFLAGS_OSX32      = $(CFLAGS) -m32 -DOSX
CFLAGS_OSX64      = $(CFLAGS) -m64 -DOSX

all: linux windows mac

.PHONY: clean
clean:
	@rm -f -R ./bin/*

linux: linux32 linux64
windows: windows32 windows64
mac: mac32 mac64

linux32: ./src/d-bf.c
	@mkdir -p ./bin
	-$(CC_LINUX32)		$(CFLAGS_LINUX32)	-o ./bin/$(MAIN_NAME)_32$(EXT_LINUX) $^

linux64: ./src/d-bf.c
	@mkdir -p ./bin
	-$(CC_LINUX64)		$(CFLAGS_LINUX64)	-o ./bin/$(MAIN_NAME)_64$(EXT_LINUX) $^

windows32: ./src/d-bf.c
	@mkdir -p ./bin
	-$(CC_WINDOWS32)	$(CFLAGS_WINDOWS32)	-o ./bin/$(MAIN_NAME)_32$(EXT_WINDOWS) $^

windows64: ./src/d-bf.c
	@mkdir -p ./bin
	-$(CC_WINDOWS64)	$(CFLAGS_WINDOWS64)	-o ./bin/$(MAIN_NAME)_64$(EXT_WINDOWS) $^

mac32: ./src/d-bf.c
	@mkdir -p ./bin
	-$(CC_OSX32)		$(CFLAGS_OSX32)		-o ./bin/$(MAIN_NAME)_32$(EXT_MAC) $^

mac64: ./src/d-bf.c
	@mkdir -p ./bin
	-$(CC_OSX64)		$(CFLAGS_OSX64)		-o ./bin/$(MAIN_NAME)_64$(EXT_MAC) $^