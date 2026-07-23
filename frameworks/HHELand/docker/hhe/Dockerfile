# syntax=docker/dockerfile:1.7

ARG GO_IMAGE=golang:1.24-bookworm@sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac
ARG TOOLING_IMAGE=hhe-arena:tooling

FROM ${GO_IMAGE} AS builder
WORKDIR /src
COPY frameworks/HHELand/go.mod frameworks/HHELand/go.sum ./
RUN go mod download
COPY frameworks/HHELand/ ./
RUN CGO_ENABLED=0 go test -c -trimpath -ldflags='-s -w -buildid=' -o /out/hheland-hera-hhe.test ./hhe/hera \
 && CGO_ENABLED=0 go test -c -trimpath -ldflags='-s -w -buildid=' -o /out/hheland-pasta-hhe.test ./hhe/pasta \
 && CGO_ENABLED=0 go test -c -trimpath -ldflags='-s -w -buildid=' -o /out/hheland-rubato-hhe.test ./hhe/rubato \
 && CGO_ENABLED=0 go test -c -trimpath -ldflags='-s -w -buildid=' -o /out/hheland-rtf-hhe.test ./rtf_integration \
 && CGO_ENABLED=0 go test -c -trimpath -ldflags='-s -w -buildid=' -o /out/hheland-pure-he.test ./he

FROM ${TOOLING_IMAGE} AS runtime
ARG IMAGE_REVISION=unknown
ARG IMAGE_VERSION=development
ARG TOOLING_DIGEST=development
LABEL org.opencontainers.image.title="HHE-ARENA HHELand HE/HHE benchmarks" \
      org.opencontainers.image.description="Prebuilt HHELand HHE, RtF, BGV, and CKKS benchmark entry points" \
      org.opencontainers.image.source="https://github.com/MarbleHE/HHE-ARENA" \
      org.opencontainers.image.revision="${IMAGE_REVISION}" \
      org.opencontainers.image.version="${IMAGE_VERSION}" \
      io.hhe-arena.contract.version="1" \
      io.hhe-arena.matrix.id="hheland-hhe-he-go" \
      io.hhe-arena.manifest.path="/opt/hhe/benchmarks/manifest.json" \
      io.hhe-arena.results.path="/work/results" \
      io.hhe-arena.resource.class="extreme" \
      io.hhe-arena.tooling.digest="${TOOLING_DIGEST}"
USER root
COPY --from=builder /out/ /opt/hhe/frameworks/hheland-hhe/
COPY manifests/hheland-hhe-he.v2.json /opt/hhe/benchmarks/manifest.json
RUN mkdir -p /work/results \
 && chown -R 65532:65532 /work \
 && hhe-validate-framework-image --require-benchmarks \
 && hhe list >/dev/null
USER 65532:65532
WORKDIR /work
ENTRYPOINT ["hhe-tooling"]
CMD ["--help"]