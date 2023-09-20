.PHONY : serve
serve : build
	APIARY_INTERFACE=localhost ./cmd/apiary/apiary

.PHONY : build
build :
	go build
	cd cmd/apiary && go build

.PHONY : install
install :
	go build
	cd cmd/apiary && go install

.PHONY : test
test :
	go test ./...

.PHONY : bench
bench: export APIARY_LOGGING=false
bench :
	go test ./... -run=^$$ -bench=.

.PHONY : vuln
vuln : 
	govulncheck ./...

.PHONY : docker-build
docker-build : 
	docker build --tag apiary:test .

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
