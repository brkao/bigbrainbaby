run: bin/bbb
	@PATH="$(PWD)/bin:$(PATH)" heroku local

bin/bbb: main.go reddit.go
	go build -o bin/bbb main.go reddit.go

clean:
	rm -rf bin

