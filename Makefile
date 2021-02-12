run: bin/bigbrainbaby
	@PATH="$(PWD)/bin:$(PATH)" heroku local

bin/bigbrainbaby: main.go reddit.go
	go build -o bin/bigbrainbaby main.go reddit.go

clean:
	rm -rf bin

