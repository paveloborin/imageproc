APP = imageproc-service
GOOS?=linux
DOCKERFILE_PATH ?=./Dockerfile
IMAGE_VERSION=1.0

build-grpc:
	protoc -I . ./proto/service.proto  --go_out=plugins=grpc:.

run-server:
	go run cmd/backend/*.go --grpc-port 8050 --log-level debug --tmp-dir ./data/tmp --py-script ./scripts/image-impr.py

run-client:
	go run cmd/frontend/*.go --grpc-port 8050 --log-level debug --grpc-host localhost --input-path ./data/input --output-path ./data/output

lint:
	go fmt $$(go list ./... | grep -v ./vendor/)
	goimports -d -w $$(find . -type f -name '*.go' -not -path './vendor/*')
	golangci-lint run

build-client:
	CGO_ENABLED=0 GOOS=${GOOS} go build -a -installsuffix cgo \
		-o ./bin/client ./cmd/frontend/*.go

build-server:
	CGO_ENABLED=0 GOOS=${GOOS} go build -a -installsuffix cgo \
		-o ./bin/server ./cmd/backend/*.go

docker-build:
	docker build -t poborin/imageproc:$(IMAGE_VERSION) -f $(DOCKERFILE_PATH) .

docker-run-server:
	docker run -p 8050:8050 --env-file ./env.list poborin/imageproc:$(IMAGE_VERSION)