all: build

build:
	go build -o bin/bbb main.go reddit.go

clean:
	rm -rf bin

