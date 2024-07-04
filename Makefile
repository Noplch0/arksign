BUILD_ENV := CGO_ENABLED=0
BUILD := `date +%FT%T%z`
TARGET_EXEC := arkSign

clean:
	rm -rf build

setup:
	mkdir -p build

all: setup build-linux build-windows build-darwin

build-linux:
	GOARCH=amd64 GOOS=linux ${BUILD_ENV} go build -o build/linux_amd64_${TARGET_EXEC} main.go

build-windows:
	GOARCH=amd64 GOOS=windows ${BUILD_ENV} go build -o build/windows_amd64_${TARGET_EXEC}.exe main.go

build-darwin:
	GOARCH=amd64 GOOS=darwin ${BUILD_ENV} go build -o build/darwin_amd64_${TARGET_EXEC} main.go

