tools:
	go -mod vendor build -o bin/build-sfomuseum-data/main.go cmd/build-sfomuseum-data.go
	go -mod vendor build -o bin/build-flysfo-data/main.go cmd/build-flysfo-data.go

data:	tools sfomuseum-data flysfo-data

sfomuseum-data:
	bin/build-sfomuseum-data > /usr/local/sfomuseum/go-sfomuseum-airlines/sfomuseum/data.go

flysfo-data:
	bin/build-flysfo-data > /usr/local/sfomuseum/go-sfomuseum-airlines/flysfo/data.go
