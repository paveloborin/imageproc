APP = imageproc-service

build:
	protoc -I . ./proto/service.proto  --go_out=plugins=grpc:.

run_server:
	go run cmd/backend/*.go --grpc-port 8050 --log-level debug --tmp-dir ./data/tmp --py-script ./scripts/image-impr.py

run_client:
	go run cmd/frontend/*.go --grpc-port 8050 --log-level debug --grpc-host localhost --input-path ./data/input --output-path ./data/output