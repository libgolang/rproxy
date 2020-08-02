build:
	go build -gcflags=all="-N -l"

install:
	go install

test:
	go test
	make -C lib test
dist: rproxy-linux-amd64 rproxy-windows-amd64.exe rproxy-darwin-amd64

rproxy-linux-amd64:
	GOARCH=amd64 GOOS=linux go build -o rproxy-linux-amd64

rproxy-windows-amd64.exe:
	GOARCH=amd64 GOOS=windows go build -o rproxy-windows-amd64.exe

rproxy-darwin-amd64:
	GOARCH=amd64 GOOS=darwin go build -o rproxy-darwin-amd64

