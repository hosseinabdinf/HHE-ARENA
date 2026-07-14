# syntax=docker/dockerfile:1.7

ARG GO_IMAGE=golang:1.24-bookworm@sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac
ARG TOOLING_IMAGE=hhe-arena:tooling
ARG LATTIGO_COMMIT=661575ca7ec52f03b135ce676994e5e8b72188dd

FROM ${GO_IMAGE} AS builder
ARG LATTIGO_COMMIT
WORKDIR /src
RUN git init /src \
 && git -C /src remote add origin https://github.com/tuneinsight/lattigo.git \
 && git -C /src fetch --depth 1 origin "${LATTIGO_COMMIT}" \
 && git -C /src checkout --detach FETCH_HEAD \
 && test "$(git -C /src rev-parse HEAD)" = "${LATTIGO_COMMIT}"
COPY frameworks/AES/BtR_framework/go.mod frameworks/AES/BtR_framework/go.sum /tmp/btr-overlay/
COPY frameworks/AES/BtR_framework/ckks_cipher/ /tmp/btr-overlay/ckks_cipher/
COPY frameworks/AES/BtR_framework/examples/btr_cipher/ /tmp/btr-overlay/examples/btr_cipher/
RUN cp /tmp/btr-overlay/go.mod /src/go.mod \
 && cp /tmp/btr-overlay/go.sum /src/go.sum \
 && mkdir -p /src/ckks_cipher /src/examples/btr_cipher \
 && cp -a /tmp/btr-overlay/ckks_cipher/. /src/ckks_cipher/ \
 && cp -a /tmp/btr-overlay/examples/btr_cipher/. /src/examples/btr_cipher/
RUN go mod download
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w -buildid=' \
      -o /out/btr-aes ./examples/btr_cipher

FROM ${TOOLING_IMAGE} AS runtime
ARG IMAGE_REVISION=unknown
ARG IMAGE_VERSION=development
ARG TOOLING_DIGEST=development
LABEL org.opencontainers.image.title="HHE-ARENA BtR AES/CKKS benchmark" \
      org.opencontainers.image.description="Prebuilt BtR AES CKKS transciphering benchmark with the HHE-ARENA contract" \
      org.opencontainers.image.source="https://github.com/MarbleHE/HHE-ARENA" \
      org.opencontainers.image.revision="${IMAGE_REVISION}" \
      org.opencontainers.image.version="${IMAGE_VERSION}" \
      io.hhe-arena.contract.version="1" \
      io.hhe-arena.matrix.id="btr-aes-ckks-go" \
      io.hhe-arena.manifest.path="/opt/hhe/benchmarks/manifest.json" \
      io.hhe-arena.results.path="/work/results" \
      io.hhe-arena.resource.class="extreme" \
      io.hhe-arena.tooling.digest="${TOOLING_DIGEST}"
USER root
COPY --from=builder /out/btr-aes /opt/hhe/frameworks/btr/btr-aes
COPY manifests/btr-aes-ckks.v2.json /opt/hhe/benchmarks/manifest.json
RUN mkdir -p /work/results \
 && chown -R 65532:65532 /work \
 && hhe-validate-framework-image --require-benchmarks \
 && hhe list >/dev/null
USER 65532:65532
WORKDIR /work
ENTRYPOINT ["hhe-tooling"]
CMD ["--help"]
