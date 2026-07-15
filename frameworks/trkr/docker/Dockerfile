# syntax=docker/dockerfile:1.7

ARG RUST_IMAGE=rust:1.85-bookworm@sha256:e51d0265072d2d9d5d320f6a44dde6b9ef13653b035098febd68cce8fa7c0bc4
ARG TOOLING_IMAGE=hhe-arena:tooling

FROM ${RUST_IMAGE} AS builder
WORKDIR /src
COPY benchmarking-utilities/rust/ benchmarking-utilities/rust/
COPY benchmarking-utilities/rust-criterion-compat/ benchmarking-utilities/rust-criterion-compat/
COPY frameworks/trkr/ frameworks/trkr/
WORKDIR /src/frameworks/trkr/apps/trivium
RUN cargo build --locked --release --benches \
 && mkdir -p /out/bin \
 && find /src/frameworks/trkr/target/release/deps -maxdepth 1 -type f -name 'trivium-*' -perm -111 -exec cp {} /out/bin/trkr_trivium \; \
 && test -x /out/bin/trkr_trivium \
 && strip /out/bin/trkr_trivium

FROM ${TOOLING_IMAGE} AS runtime
ARG IMAGE_REVISION=unknown
ARG IMAGE_VERSION=development
ARG TOOLING_DIGEST=development
LABEL org.opencontainers.image.title="HHE-ARENA trkr TFHE-rs benchmarks" \
      org.opencontainers.image.description="Prebuilt trkr Trivium and Kreyvium Rust benchmarks with the HHE-ARENA contract" \
      org.opencontainers.image.source="https://github.com/MarbleHE/HHE-ARENA" \
      org.opencontainers.image.revision="${IMAGE_REVISION}" \
      org.opencontainers.image.version="${IMAGE_VERSION}" \
      io.hhe-arena.contract.version="1" \
      io.hhe-arena.matrix.id="trkr-tfhe-rust" \
      io.hhe-arena.manifest.path="/opt/hhe/benchmarks/manifest.json" \
      io.hhe-arena.results.path="/work/results" \
      io.hhe-arena.resource.class="extreme" \
      io.hhe-arena.tooling.digest="${TOOLING_DIGEST}"
USER root
COPY --from=builder /out/bin/ /opt/hhe/frameworks/trkr/
COPY manifests/trkr.v2.json /opt/hhe/benchmarks/manifest.json
RUN mkdir -p /work/results \
 && chown -R 65532:65532 /work \
 && hhe-validate-framework-image --require-benchmarks \
 && hhe list >/dev/null
USER 65532:65532
WORKDIR /work
ENTRYPOINT ["hhe-tooling"]
CMD ["--help"]
