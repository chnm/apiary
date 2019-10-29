serve : build
	source .env && ./relecapi

build :
	cd ../.. && go build
	go build

install :
	cd ../.. && go build
	go install

test :
	source .env && go test -v
