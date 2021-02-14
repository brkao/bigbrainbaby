run: bigbrainbaby
	@PATH="$(PWD):$(PATH)" heroku local web

bigbrainbaby: main.go
	go build -o bigbrainbaby main.go

clean:
	rm -rf bin

