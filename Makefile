ifndef version
	version=latest
endif

default: linux

image:
	@echo "building image..."
	docker build -t agile:$(version) .

linux:
	GOOS=linux GOARCH=amd64 go build -o agile
