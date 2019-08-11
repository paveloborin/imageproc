# Image Processing Service

Simple GRPC-service for removing text from small images and image color auto correction.

Example:

![alt text](data/output/1.jpg)

## Usage

First, run server via prepared docker image by command `make docker-run-server`, or install python3, cv2 library and run `make run-server` command,

Build client by command `make build-client`.

Run `./bin/client --grpc-port 8050 --grpc-host localhost --input-path ./data/input --output-path ./data/output`,
where an input-path is your images folder, output-path - the folder with result images.

