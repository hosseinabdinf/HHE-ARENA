# syntax=docker/dockerfile:1.7

ARG RUST_IMAGE=rust:1.85-bookworm@sha256:e51d0265072d2d9d5d320f6a44dde6b9ef13653b035098febd68cce8fa7c0bc4
ARG TOOLING_IMAGE=hhe-arena:tooling

FROM ${RUST_IMAGE} AS builder
WORKDIR /src
COPY benchmarking-utilities/rust/ benchmarking-utilities/rust/
COPY frameworks/AES/P_encoding/ frameworks/AES/P_encoding/
WORKDIR /src/frameworks/AES/P_encoding/crypto_primitives
RUN cargo build --release \
 && mkdir -p /out/bin \
 && cp target/release/crypto_primitives /out/bin/run-aes \
 && strip /out/bin/run-aes \
 && test -x /out/bin/run-aes

FROM ${TOOLING_IMAGE} AS runtime
ARG IMAGE_REVISION=unknown
ARG IMAGE_VERSION=development
ARG TOOLING_DIGEST=development
LABEL org.opencontainers.image.title="HHE-ARENA P-encoding AES/Rust benchmarks" \
      org.opencontainers.image.description="Prebuilt P-encoding odd-modulus TFHE AES benchmark with the HHE-ARENA contract" \
      org.opencontainers.image.source="https://github.com/MarbleHE/HHE-ARENA" \
      org.opencontainers.image.revision="${IMAGE_REVISION}" \
      org.opencontainers.image.version="${IMAGE_VERSION}" \
      io.hhe-arena.contract.version="1" \
      io.hhe-arena.matrix.id="p-encoding-aes-rust" \
      io.hhe-arena.manifest.path="/opt/hhe/benchmarks/manifest.json" \
      io.hhe-arena.results.path="/work/results" \
      io.hhe-arena.resource.class="extreme" \
      io.hhe-arena.tooling.digest="${TOOLING_DIGEST}"
USER root
COPY --from=builder /out/bin/run-aes /opt/hhe/frameworks/p-encoding/run-aes
COPY manifests/p-encoding-aes.v2.json /opt/hhe/benchmarks/manifest.json
RUN mkdir -p /work/results \
 && chown -R 65532:65532 /work \
 && hhe-validate-framework-image --require-benchmarks \
 && hhe list >/dev/null
USER 65532:65532
WORKDIR /work
ENTRYPOINT ["hhe-tooling"]
CMD ["--help"]
