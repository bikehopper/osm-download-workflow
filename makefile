install:
	go mod download && go mod verify
build:
	go build -o ./bin/osm-download-workflow ./src/cmd
build_docker: 
	docker build . -t osm-download-workflow:local
run:
	go run ./src/cmd
all: install build