build:
	go build

build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build -o slr-linux-amd64

build-and-scp:
ifndef TO
	$(error TO is undefined, use 'make build-and-scp TO=root@server.example.com:/slu')
endif

	@make build-linux-amd64
	scp slr-linux-amd64 ${TO}

release:
	slu go-code version-bump --auto --tag
	slu go-code version-bump --auto
	git push
	git push --tags
