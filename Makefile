testserver:
	go run -ldflags="-X 'main.VERSION=test'" ./cmd/lcode-hub
build:
	go build -ldflags="-X 'main.VERSION=$$(git describe --tags --always --dirty | cut -c2-)' -s -w" -o lcode-hub ./cmd/lcode-hub
build-with-upx: build
	upx lcode-hub
