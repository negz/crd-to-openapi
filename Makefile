.PHONY: all
all: build


.PHONY: build
build:
	CGO_ENABLED=0 go build -mod vendor -buildmode=pie -a -o bin/crd-to-openapi cmd/main.go

.PHONY: build-multi
build-multi:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -mod vendor -buildmode=pie -a -o bin/crd-to-openapi.linux-arm64 cmd/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod vendor -buildmode=pie -a -o bin/crd-to-openapi.linux-amd64 cmd/main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -mod vendor -buildmode=pie -a -o bin/crd-to-openapi.darwin-arm64 cmd/main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -mod vendor -buildmode=pie -a -o bin/crd-to-openapi.darwin-amd64 cmd/main.go
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -mod vendor -buildmode=pie -a -o bin/crd-to-openapi.windows-amd64 cmd/main.go