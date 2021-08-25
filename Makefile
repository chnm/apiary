GITBRANCH := $(shell git branch --show-current)

# Run Docker image associated with branch from GitHub Container Registry
.PHONY : serve-ghcr
serve-ghcr :
	docker pull ghcr.io/chnm/dataapi:$(GITBRANCH)
	docker run --rm \
		--publish 8090:8090 \
		-e DATAAPI_DBHOST \
		-e DATAAPI_DBPORT \
		-e DATAAPI_DBPASS \
		-e DATAAPI_DBUSER \
		-e DATAAPI_DBNAME \
		-e DATAAPI_APB \
		--name dataapi-dev \
		ghcr.io/chnm/dataapi:$(GITBRANCH)

