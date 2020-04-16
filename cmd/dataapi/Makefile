serve : build
	./dataapi

build :
	cd ../.. && go build
	go build

install :
	cd ../.. && go build
	go install

test :
	go test -v

docker-build : 
	docker build --tag dataapi:test ../..

# This assumes that the environment variables are available
docker-serve : docker-build
	docker run --rm \
		--publish 8090:8090 \
		-e DATAAPI_DBHOST \
		-e DATAAPI_DBPORT \
		-e DATAAPI_DBNAME \
		-e DATAAPI_DBPASS \
		-e DATAAPI_DBUSER \
		--name dataapi \
		dataapi:test

