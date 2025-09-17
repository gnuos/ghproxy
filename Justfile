default:
    @just build

build:
    CGO_ENABLE=0 go build -buildvcs=false -gcflags="all=-N -l" -ldflags='-w -s -buildid=' -trimpath -o bin/ .

clean:
    rm -f ./bin/*

