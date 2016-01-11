all: build
version=1.0.0
COMPILE=$(shell date -u "+%Y-%m-%d/%H:%M:%S")
build:
	go build -ldflags "-X github.com/lijiaocn/GoPkgs/version.VERSION ${version} -X github.com/lijiaocn/GoPkgs/version.COMPILE ${COMPILE}"  

