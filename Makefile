BINARY  := gogoversion
LINK    := ggv
INSTALL := $(shell go env GOPATH)/bin

.PHONY: build install uninstall clean run tidy

build:
	go build -o $(BINARY) .

tidy:
	go mod tidy

install: build
	cp $(BINARY) $(INSTALL)/$(BINARY)
	ln -sf $(INSTALL)/$(BINARY) $(INSTALL)/$(LINK)

uninstall:
	rm -f $(INSTALL)/$(BINARY) $(INSTALL)/$(LINK)

clean:
	rm -f $(BINARY)

run:
	go run . --dry-run