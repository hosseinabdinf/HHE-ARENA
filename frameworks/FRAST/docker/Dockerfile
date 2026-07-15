# syntax=docker/dockerfile:1.7

ARG RUST_IMAGE=rust:1.85-bookworm@sha256:e51d0265072d2d9d5d320f6a44dde6b9ef13653b035098febd68cce8fa7c0bc4
ARG TOOLING_IMAGE=hhe-arena:tooling

FROM ${RUST_IMAGE} AS builder
WORKDIR /src
COPY benchmarking-utilities/rust/ benchmarking-utilities/rust/
COPY benchmarking-utilities/rust-criterion-compat/ benchmarking-utilities/rust-criterion-compat/
COPY frameworks/FRAST/ frameworks/FRAST/
WORKDIR /src/frameworks/FRAST/frast
RUN cargo build --locked --release --tests --benches \
 && mkdir -p /out/bin \
 && find target/release/deps -maxdepth 1 -type f -name 'frast-*' -perm -111 -exec cp {} /out/bin/frast_test \; \
 && find target/release/deps -maxdepth 1 -type f -name 'bench_setup-*' -perm -111 -exec cp {} /out/bin/bench_setup \; \
 && find target/release/deps -maxdepth 1 -type f -name 'bench_frast-*' -perm -111 -exec cp {} /out/bin/bench_frast \; \
 && find target/release/deps -maxdepth 1 -type f -name 'bench_gen_pbs-*' -perm -111 -exec cp {} /out/bin/bench_gen_pbs \; \
 && find target/release/deps -maxdepth 1 -type f -name 'bench_online-*' -perm -111 -exec cp {} /out/bin/bench_online \; \
 && test -x /out/bin/frast_test \
 && test -x /out/bin/bench_setup \
 && test -x /out/bin/bench_frast \
 && test -x /out/bin/bench_gen_pbs \
 && test -x /out/bin/bench_online \
 && strip /out/bin/*

FROM ${TOOLING_IMAGE} AS runtime
ARG IMAGE_REVISION=unknown
ARG IMAGE_VERSION=development
ARG TOOLING_DIGEST=development
LABEL org.opencontainers.image.title="HHE-ARENA FRAST Rust benchmarks" \
      org.opencontainers.image.description="Prebuilt FRAST Rust benchmarks with the HHE-ARENA contract" \
      org.opencontainers.image.source="https://github.com/MarbleHE/HHE-ARENA" \
      org.opencontainers.image.revision="${IMAGE_REVISION}" \
      org.opencontainers.image.version="${IMAGE_VERSION}" \
      io.hhe-arena.contract.version="1" \
      io.hhe-arena.matrix.id="frast-rust" \
      io.hhe-arena.manifest.path="/opt/hhe/benchmarks/manifest.json" \
      io.hhe-arena.results.path="/work/results" \
      io.hhe-arena.resource.class="extreme" \
      io.hhe-arena.tooling.digest="${TOOLING_DIGEST}"
USER root
COPY --from=builder /out/bin/ /opt/hhe/frameworks/frast/
COPY manifests/frast.v2.json /opt/hhe/benchmarks/manifest.json
RUN mkdir -p /work/results \
 && chown -R 65532:65532 /work \
 && hhe-validate-framework-image --require-benchmarks \
 && hhe list >/dev/null
USER 65532:65532
WORKDIR /work
ENTRYPOINT ["hhe-tooling"]
CMD ["--help"]
