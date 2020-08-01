
build-debug:
	go build -gcflags=all="-N -l"
build:
	go build

install:
	go install

test:
	go test
	make -C lib test
