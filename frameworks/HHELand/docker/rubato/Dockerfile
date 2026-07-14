# syntax=docker/dockerfile:1.7

ARG GO_IMAGE=golang:1.24-bookworm@sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac
ARG TOOLING_IMAGE=hhe-arena:tooling

FROM ${GO_IMAGE} AS builder
WORKDIR /src
COPY frameworks/HHELand/go.mod frameworks/HHELand/go.sum ./
RUN go mod download
COPY frameworks/HHELand/ ./
RUN CGO_ENABLED=0 go test -c \
      -trimpath \
      -ldflags='-s -w -buildid=' \
      -o /out/hheland-rubato.test \
      ./sym/rubato

FROM ${TOOLING_IMAGE} AS runtime
ARG IMAGE_REVISION=unknown
ARG IMAGE_VERSION=development
ARG TOOLING_DIGEST=development
LABEL org.opencontainers.image.title="HHE-ARENA HHELand/Rubato benchmark" \
      org.opencontainers.image.description="Prebuilt Rubato-128-L Go symmetric benchmark with the HHE-ARENA contract" \
      org.opencontainers.image.source="https://github.com/MarbleHE/HHE-ARENA" \
      org.opencontainers.image.revision="${IMAGE_REVISION}" \
      org.opencontainers.image.version="${IMAGE_VERSION}" \
      io.hhe-arena.contract.version="1" \
      io.hhe-arena.matrix.id="hheland-rubato128l-go" \
      io.hhe-arena.manifest.path="/opt/hhe/benchmarks/manifest.json" \
      io.hhe-arena.results.path="/work/results" \
      io.hhe-arena.resource.class="light" \
      io.hhe-arena.tooling.digest="${TOOLING_DIGEST}"
USER root
COPY --from=builder /out/hheland-rubato.test /opt/hhe/frameworks/hheland-rubato/hheland-rubato.test
COPY manifests/hheland-rubato.v1.json /opt/hhe/benchmarks/manifest.json
RUN mkdir -p /work/results \
 && chown -R 65532:65532 /work \
 && hhe-validate-framework-image --require-benchmarks \
 && hhe list >/dev/null
USER 65532:65532
WORKDIR /work
ENTRYPOINT ["hhe-tooling"]
CMD ["--help"]
