
lint:
	golangci-lint run

build: 
	go build -ldflags "-X main.buildCommit=${shell git rev-parse HEAD}" cmd/main.go
		
	