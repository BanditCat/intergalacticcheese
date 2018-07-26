export PATH:=/usr/bin:$(PATH)

TARGET=intergalacticcheese.exe

TXTS:=$(TXTS) $(wildcard ./*.txt) ./Makefile ./README.md ./windowsResource.rc
SRCS:=$(SRCS) $(wildcard ./*.go)

$(TARGET):
	go build $(GOBUILDFLAGS)
	$(RESCOMPILE)
	$(RESATTACH)
	$(COMPACT)
	$(RESCLEAN)

.PHONY: release 
release: $(TARGET) windowsResource.rc
release: GOBUILDFLAGS:=-ldflags '-s -w'
release: COMPACT:=upx --best $(TARGET)
release: RESCOMPILE:=ResourceHacker -open windowsResource.rc -save win.res -action compile
release: RESATTACH:=ResourceHacker -open $(TARGET) -save $(TARGET) -action addoverwrite -res win.res
release: RESCLEAN:=/usr/bin/rm win.res

.PHONY: clean
clean:
	go clean

.PHONY: backup
backup: clean release
	git add -A
	git commit -a -F gitmessage.txt || true

.PHONY: run
run: $(TARGET)
	./$(TARGET)
