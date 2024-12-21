.PHONY: build run clean test

build: clean
	go build -o fp main.go

run: build
	./fp

clean:
	rm -f fp

test:
	go test -v ./...

run-services:
	docker compose up -d

stop-services:
	docker compose down
