FROM --platform=$BUILDPLATFORM golang:1.27.2-bookworm AS build
WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags="-s -w" -o /zio-learn ./cmd/zio-learn

FROM eclipse-temurin:21-jdk-jammy
ARG SCALA_CLI_VERSION=1.18.0
RUN apt-get update && apt-get install -y --no-install-recommends curl ca-certificates jq \
    && rm -rf /var/lib/apt/lists/* \
    && mkdir -p /opt/scala-cli \
    && curl -fSL --retry 3 "https://repo.maven.apache.org/maven2/org/virtuslab/scala-cli/cliBootstrapped/${SCALA_CLI_VERSION}/cliBootstrapped-${SCALA_CLI_VERSION}.jar" -o /opt/scala-cli/scala-cli.jar \
    && printf '#!/bin/sh\nexec java -jar /opt/scala-cli/scala-cli.jar "$@"\n' > /usr/local/bin/scala-cli \
    && chmod +x /usr/local/bin/scala-cli \
    && useradd --create-home --uid 10001 playground
ENV COURSIER_REPOSITORIES=https://repo.maven.apache.org/maven2 \
    LISTEN_ADDR=0.0.0.0:8080
USER playground
WORKDIR /home/playground
# Cache dependencies before copying the server, so Go changes reuse this layer.
COPY internal/assets/config.json /tmp/scala-config.json
COPY docker/cache-scala.sh /tmp/cache-scala.sh
RUN sh /tmp/cache-scala.sh /tmp/scala-config.json
COPY --from=build /zio-learn /usr/local/bin/zio-learn
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s CMD curl -fsS http://127.0.0.1:8080/healthz || exit 1
ENTRYPOINT ["zio-learn"]
