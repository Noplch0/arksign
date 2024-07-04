BUILD_ENV := CGO_ENABLED=0
BUILD=`date +%FT%T%z`
TARGET_EXEC := arkSign

clean:
	rm -rf build

setup:
	mkdir -p /build

all:setup build-linux build-windows build-darwin

build-linux:
	${BUILD_ENV} GOARCH=amd64 GOOS=linux go build main.go -o build/${GOOS}_${GOARCH}_${TARGET_EXEC}

build-windows:
	${BUILD_ENV} GOARCH=amd64 GOOS=windows go build main.go -o build/${GOOS}_${GOARCH}_${TARGET_EXEC}

budil-darwin:
	${BUILD_ENV} GOARCH=amd64 GOOS=darwin go build main.go -o build/${GOOS}_${GOARCH}_${TARGET_EXEC}
