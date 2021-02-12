run: bin/bbb
	@PATH="$(PWD)/bin:$(PATH)" heroku local

bin/bbb:
	go build -o bin/bbb main.go reddit.go

clean:
	rm -rf bin

