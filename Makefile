#.PHONY: lint test build docker push deploy all
.PHONY: lint build docker push all


MAKEFILE_PATH=$(shell readlink -f "${0}")
MAKEFILE_DIR=$(shell dirname "${MAKEFILE_PATH}")

version=$(shell grep 'image: adiazny/easy-strava-upload:' deployments/kubernetes/deployment.yml | awk -F: '{print $$3}')

parentImage=alpine:latest

lint:
	golangci-lint run ./...

#test:
#	go test -v -race -coverprofile=coverage.out ./...

build:
	CGO_ENABLED=0 go build -o build/package/easy-strava-upload cmd/easy-strava-upload/easy-strava-upload.go

image:
	docker pull "${parentImage}"
	docker image build -t adiazny/easy-strava-uplaod:${version} build/package/easy-strava-upload

push:
	docker login -u "${DOCKER_USER}" -p "${DOCKER_PASS}"
	docker push adiazny/easy-strava-uplaod:${version}
	docker tag adiazny/easy-strava-uplaod:${version} adiazny/easy-strava-uplaod:latest
	docker logout

deploy:
	${MAKEFILE_DIR}/scripts/deploy.sh

#all: lint test build image push deploy
all: lint build image push