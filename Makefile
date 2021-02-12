all: build

build:
	go build -o bbb main.go reddit.go

run:
	./bbb

