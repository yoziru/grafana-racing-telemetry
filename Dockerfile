# syntax=docker/dockerfile:1.3
FROM node:lts-alpine AS yarn-builder
ENV YARN_CACHE_FOLDER=/opt/yarncache

WORKDIR /app/simracing-telemetry-datasource

# Install yarn dependencies.
COPY ./package.json ./yarn.lock /app/simracing-telemetry-datasource/
RUN yarn install --frozen-lockfile --network-timeout 150000 

# Build plugin.
COPY ./src /app/simracing-telemetry-datasource/src
COPY README.md CHANGELOG.md package.json jest.config.js tsconfig.json .prettierrc.js yarn.lock LICENSE \
    /app/simracing-telemetry-datasource/
RUN yarn build

FROM golang:alpine AS mage-builder
ENV GOCACHE=/opt/go/gocache \
    GOMODCACHE=/opt/go/gomodcache

RUN apk add git
RUN git clone https://github.com/magefile/mage -b v1.11.0 --single-branch \
    && cd mage \
    && go run bootstrap.go
WORKDIR /app/simracing-telemetry-datasource

COPY go.mod go.sum /app/simracing-telemetry-datasource/
RUN --mount=type=cache,target=/opt/go \
    go mod download

COPY . /app/simracing-telemetry-datasource
RUN --mount=type=cache,target=/opt/go \
    if [ $(uname -m) == aarch64 ]; then \
    mage -v build:linuxarm64; \
    elif [ $(uname -m) == armv7l ]; then \
    mage -v build:linuxarm; \
    else \
    mage -v build:linux; \
    fi

FROM grafana/grafana:latest

ENV GF_DEFAULT_APP_MODE development

# Copy plugin files into custom location, to avoid conflicting with contents of /var/lib/grafana. Point
# Grafana to this directory as additional plugin path with the GF_PATHS_PLUGINS env var.
ARG TARGETARCH
# set default env
ENV GF_DEFAULT_APP_MODE=development \
    GF_AUTH_ANONYMOUS_ENABLED=true \
    GF_AUTH_ANONYMOUS_ORG_ROLE="Admin" \
    GF_SERVER_ROOT_URL=http://admin:admin@grafana:3000 \
    GF_LOG_LEVEL=debug \
    GF_LOG_FILTERS=plugins:debug \
    GF_DATAPROXY_LOGGING=true \
    DEV_ALLOW_EMPTY_API_KEY=true \
    GF_INSTALL_PLUGINS=cloudspout-button-panel,https://github.com/yoziru/yoziru-3dseries-panel/releases/download/v1.0.0/yoziru-3dseries-panel-1.0.0.zip;yoziru-3dseries-panel \
    CUSTOM_PLUGIN_DIR=/home/grafana/plugins
RUN mkdir -p ${CUSTOM_PLUGIN_DIR}
# RUN echo '{"apiVersion":1,"apps":[{"type":"yoziru-3dseries-panel"}]}' \
#     > /etc/grafana/provisioning/plugins/plugins.yaml
COPY --from=yarn-builder /app/simracing-telemetry-datasource/dist ${CUSTOM_PLUGIN_DIR}/simracing-telemetry-datasource/dist
COPY --from=mage-builder /app/simracing-telemetry-datasource/dist/gpx_simracing-telemetry-datasource_linux_${TARGETARCH} \
    ${CUSTOM_PLUGIN_DIR}/simracing-telemetry-datasource/dist/gpx_simracing-telemetry-datasource_linux_${TARGETARCH}
ENV GF_PATHS_PLUGINS ${CUSTOM_PLUGIN_DIR}

# Enable plugin by default (but is not configured)
RUN echo '{"apiVersion":1,"datasources":[{"name":"simracing-telemetry-datasource","type":"grafana-simracing-telemetry-datasource", "jsonData":{"recordingBasePath": "/var/lib/grafana/simracing-telemetry","recordingBufferDataPoints": 3600}, "isDefault": true, "editable": true}]}' \
    > /etc/grafana/provisioning/datasources/datasources.yaml
