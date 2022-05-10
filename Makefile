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

