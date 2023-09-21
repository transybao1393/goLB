BINARY_NAME = load-balancer
build:
	go build -o bin/$(BINARY_NAME) -v

build-and-run:
	go build -o bin/$(BINARY_NAME) -v
	./bin/$(BINARY_NAME)

setup:
	# create config.json
	# create bin/ folder