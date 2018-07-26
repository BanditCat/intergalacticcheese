TARGET=intergalacticcheese.exe

TXTS:=$(TXTS) $(wildcard ./*.txt) ./Makefile ./README.md ./windowsResource.rc
SRCS:=$(SRCS) $(wildcard ./*.go)

$(TARGET):
	go build

.PHONY: release 
release: $(TARGET)


.PHONY: clean
clean:
	go clean

.PHONY: backup
backup: clean release
	git add -A
	git commit -a -m "$(shell cat ./gitmessage.txt)" || true

.PHONY: run
run: $(TARGET)
	./$(TARGET)
