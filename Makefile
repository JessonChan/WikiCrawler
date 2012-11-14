all: build
		./crawler

clean:
		go clean

build: clean
		go build crawler.go store.go
