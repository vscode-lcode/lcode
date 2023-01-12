testserver:
	go run -ldflags="-X 'main.VERSION=test'" ./cmd/lcode-hub
build:
	go build -ldflags="-X 'main.VERSION=build'" -ldflags="-s -w" -o lcode-hub ./cmd/lcode-hub
	upx lcode-hub
