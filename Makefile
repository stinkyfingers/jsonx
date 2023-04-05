.PHONY: build

build:
	go build -o jsonx .
	
.PHONY: install

install: build 
	mv jsonx /usr/local/bin