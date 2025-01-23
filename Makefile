ifndef version
	version=latest
endif

default: image

image:
	@echo "building image..."
	docker build -t agile:$(version) .
