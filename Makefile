cli:
	go build -mod vendor  -o bin/build-sfomuseum-data cmd/build-sfomuseum-data/main.go
	go build -mod vendor -o bin/build-flysfo-data cmd/build-flysfo-data/main.go

data:	tools sfomuseum-data flysfo-data

sfomuseum-data:
	bin/build-sfomuseum-data > /usr/local/sfomuseum/go-sfomuseum-airlines/sfomuseum/data.go

flysfo-data:
	bin/build-flysfo-data > /usr/local/sfomuseum/go-sfomuseum-airlines/flysfo/data.go
