BIN := udirstats

.PHONY: build test clean install

build:
	CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o $(BIN) .

test:
	CGO_ENABLED=0 go test ./...

install: build
	install -Dm755 $(BIN) $(HOME)/.local/bin/$(BIN)

clean:
	rm -f $(BIN)
