.PHONY : serve
serve : build
	APIARY_INTERFACE=localhost ./apiary

.PHONY : build
build :
	cd ../.. && go build
	go build

.PHONY : install
install :
	cd ../.. && go build
	go install

.PHONY : test
test :
	go test -v

.PHONY : docker-build
docker-build : 
	docker build --tag apiary:test ../..

# This assumes that the environment variables are available
.PHONY : docker-serve
docker-serve : docker-build
	docker run --rm \
		--publish 8090:8090 \
		-e APIARY_DB \
		-e APIARY_PORT \
		-e APIARY_INTERFACE \
		-e APIARY_LOGGING \
		--name apiary \
		apiary:test


GITBRANCH := $(shell git branch --show-current)

# Run Docker image associated with branch from GitHub Container Registry
.PHONY : serve-ghcr
serve-ghcr :
	docker pull ghcr.io/chnm/apiary:$(GITBRANCH)
	docker run --rm \
		--publish 8090:8090 \
		-e APIARY_DB \
		--name apiary-dev \
		ghcr.io/chnm/apiary:$(GITBRANCH)
