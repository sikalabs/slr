-include Makefile.local.mk

build:
	go build

build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build -o slr-linux-amd64

build-and-scp:
ifndef TO
	$(error TO is undefined, use 'make build-and-scp TO=root@server.example.com:/slr')
endif

	@make build-linux-amd64
	scp slr-linux-amd64 ${TO}

release:
	git pull
	slu go-code version-bump --auto --tag
	slu go-code version-bump --auto
	git push
	git push --tags
