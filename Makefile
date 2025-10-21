build:
	go build

build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build -o slr-linux-amd64
