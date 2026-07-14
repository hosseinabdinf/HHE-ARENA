# syntax=docker/dockerfile:1.7

ARG GCC_IMAGE=gcc:14.2-bookworm@sha256:b99b86a28812b1e6453a231a947dc43d76fe192788a12f344a9b568bf9f5d24c
ARG TOOLING_IMAGE=hhe-arena:tooling

FROM ${GCC_IMAGE} AS builder
ARG DEBIAN_FRONTEND=noninteractive
ARG CMAKE_VERSION=3.25.1-1
RUN apt-get update \
 && apt-get install -y --no-install-recommends cmake=${CMAKE_VERSION} \
 && rm -rf /var/lib/apt/lists/*
WORKDIR /src
COPY frameworks/AES/Tinyaes/ ./
RUN cmake -S . -B /build \
      -DCMAKE_BUILD_TYPE=Release \
      -DCMAKE_C_FLAGS_RELEASE="-O3 -DNDEBUG" \
 && cmake --build /build --target aes_encrypt --parallel 2 \
 && install -D -m 0555 /build/aes_encrypt /out/aes_encrypt

FROM ${TOOLING_IMAGE} AS runtime
ARG IMAGE_REVISION=unknown
ARG IMAGE_VERSION=development
ARG TOOLING_DIGEST=development
LABEL org.opencontainers.image.title="HHE-ARENA AES/TinyAES benchmark" \
      org.opencontainers.image.description="Prebuilt TinyAES AES-128-CBC C benchmark with the HHE-ARENA contract" \
      org.opencontainers.image.source="https://github.com/MarbleHE/HHE-ARENA" \
      org.opencontainers.image.revision="${IMAGE_REVISION}" \
      org.opencontainers.image.version="${IMAGE_VERSION}" \
      io.hhe-arena.contract.version="1" \
      io.hhe-arena.matrix.id="tinyaes-aes128-cbc-c" \
      io.hhe-arena.manifest.path="/opt/hhe/benchmarks/manifest.json" \
      io.hhe-arena.results.path="/work/results" \
      io.hhe-arena.resource.class="light" \
      io.hhe-arena.tooling.digest="${TOOLING_DIGEST}"
USER root
COPY --from=builder /out/aes_encrypt /opt/hhe/frameworks/tinyaes/aes_encrypt
COPY manifests/tinyaes-aes.v2.json /opt/hhe/benchmarks/manifest.json
COPY schemas/benchmark-manifest.schema.json /opt/hhe/schemas/benchmark-manifest.schema.json
COPY docker/validate_framework_image.py /opt/hhe/bin/hhe-validate-framework-image
RUN chmod 0555 /opt/hhe/bin/hhe-validate-framework-image \
 && mkdir -p /work/results \
 && chown -R 65532:65532 /work \
 && hhe-validate-framework-image --require-benchmarks \
 && hhe list >/dev/null
USER 65532:65532
WORKDIR /work
ENTRYPOINT ["hhe-tooling"]
CMD ["--help"]
