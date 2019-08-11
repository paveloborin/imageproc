APP = imageproc-service
GOOS?=linux

build_grpc:
	protoc -I . ./proto/service.proto  --go_out=plugins=grpc:.

run_server:
	go run cmd/backend/*.go --grpc-port 8050 --log-level debug --tmp-dir ./data/tmp --py-script ./scripts/image-impr.py

run_client:
	go run cmd/frontend/*.go --grpc-port 8050 --log-level debug --grpc-host localhost --input-path ./data/input --output-path ./data/output

lint:
	go fmt $$(go list ./... | grep -v ./vendor/)
	goimports -d -w $$(find . -type f -name '*.go' -not -path './vendor/*')
	golangci-lint run

build_client:
	CGO_ENABLED=0 GOOS=${GOOS} go build -a -installsuffix cgo \
		-o ./bin/client ./cmd/frontend/*.go

build_server:
	CGO_ENABLED=0 GOOS=${GOOS} go build -a -installsuffix cgo \
		-o ./bin/server ./cmd/backend/*.go