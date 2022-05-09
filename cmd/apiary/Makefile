serve : build
	APIARY_INTERFACE=localhost ./apiary

build :
	cd ../.. && go build
	go build

install :
	cd ../.. && go build
	go install

test :
	go test -v

docker-build : 
	docker build --tag apiary:test ../..

# This assumes that the environment variables are available
docker-serve : docker-build
	docker run --rm \
		--publish 8090:8090 \
		-e APIARY_DB \
		-e APIARY_PORT \
		-e APIARY_INTERFACE \
		-e APIARY_LOGGING \
		--name apiary \
		apiary:test

