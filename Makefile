SHELL := /bin/bash

# ==============================================================================
# Testing running system

# Access metrics directly (4000)
# expvarmon -ports=":4000" -vars="build,requests,goroutines,errors,panics,mem:memstats.Alloc"

# ==============================================================================

run:
	go run app/services/easy-strava-upload-api/main.go

# ==============================================================================
# Building containers

VERSION := 1.0

all: strava-api

strava-api:
	docker build \
		-f zarf/docker/dockerfile.strava-api \
		-t easy-strava-upload-api-amd64:$(VERSION) \
		--build-arg BUILD_REF=$(VERSION) \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.

