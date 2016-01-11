all: build

version=1.0.0
COMPILE=$(shell date -u "+%Y-%m-%d/%H:%M:%S")
build:
	@mkdir -p _output
	godep go build -ldflags "-X github.com/lijiaocn/GoPkgs/version.VERSION=${version} -X github.com/lijiaocn/GoPkgs/version.COMPILE=${COMPILE}" -o _output/d-redis-port ./

clean:
	rm -rf _output

gotest:
	godep go test -cover -v ./...
