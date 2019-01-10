#Basic makefile

default: build

build: clean vet
	@go build -o panteon-go

doc:
	@godoc -http=:6060 -index

lint:
	@golint ./...

debug:
	@fresh

run: build
	./panteon-go

test:
	@go test ./...

vet:
	@go vet ./...

clean:
	@rm -f ./panteon-go
