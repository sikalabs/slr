-include Makefile.local.mk

build:
	go build

build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build -o slr-linux-amd64

scp-to:
ifndef TO
	$(error TO is undefined, use 'make build-and-scp TO=root@server.example.com')
endif
	scp slr-linux-amd64 ${TO}:/usr/local/bin/slr.tmp
	ssh ${TO} mv /usr/local/bin/slr.tmp /usr/local/bin/slr

build-and-scp:
	@make build-linux-amd64
	@make scp-to TO=${TO}

build-and-scp-labs:
	@make build-linux-amd64
	@make scp-to TO=root@lab0.sikademo.com
	@make scp-to TO=root@lab1.sikademo.com
	@make scp-to TO=root@lab2.sikademo.com
	@make scp-to TO=root@lab3.sikademo.com
	@make scp-to TO=root@lab4.sikademo.com
	@make scp-to TO=root@lab5.sikademo.com
	@make scp-to TO=root@lab6.sikademo.com
	@make scp-to TO=root@lab7.sikademo.com
	@make scp-to TO=root@lab8.sikademo.com
	@make scp-to TO=root@lab9.sikademo.com
	@make scp-to TO=root@lab10.sikademo.com
	@make scp-to TO=root@lab11.sikademo.com
	@make scp-to TO=root@lab12.sikademo.com
	@make scp-to TO=root@lab13.sikademo.com
	@make scp-to TO=root@lab14.sikademo.com
	@make scp-to TO=root@lab15.sikademo.com
	@make scp-to TO=root@lab16.sikademo.com
	@make scp-to TO=root@lab17.sikademo.com
	@make scp-to TO=root@lab18.sikademo.com
	@make scp-to TO=root@lab19.sikademo.com
	@make scp-to TO=root@lab20.sikademo.com

release:
	git pull
	slu go-code version-bump --auto --tag
	slu go-code version-bump --auto
	git push
	git push --tags
