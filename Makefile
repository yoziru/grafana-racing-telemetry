DOCKER_TAG := yoziru/grafana-racing-telemetry:dev
# persist data in /var/lib/grafana
VOLUMES = -v "var-lib-grafana:/var/lib/grafana"
PORTS = -p 3000:3000 \
		-p 20777:20777/udp

build:
	docker build -t $(DOCKER_TAG) .

run: build
	docker run --rm $(PORTS) $(VOLUMES) $(DOCKER_TAG) 

sh: build
	docker run --rm -it --entrypoint="" $(DOCKER_TAG) /bin/sh

push:
	docker buildx build \
	--push \
	--platform linux/arm/v7,linux/arm64/v8,linux/amd64 \
	--tag $(DOCKER_TAG) .
