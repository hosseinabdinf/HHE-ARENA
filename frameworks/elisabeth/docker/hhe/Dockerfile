# syntax=docker/dockerfile:1.7

ARG RUST_IMAGE=rust:1.85-bookworm@sha256:e51d0265072d2d9d5d320f6a44dde6b9ef13653b035098febd68cce8fa7c0bc4
ARG TOOLING_IMAGE=hhe-arena:tooling

FROM ${RUST_IMAGE} AS builder
ARG FFTW_VERSION=3.3.10-1
RUN apt-get update \
 && apt-get install -y --no-install-recommends libfftw3-dev=${FFTW_VERSION} \
 && rm -rf /var/lib/apt/lists/*
WORKDIR /src
COPY benchmarking-utilities/rust/ benchmarking-utilities/rust/
COPY frameworks/elisabeth/ frameworks/elisabeth/
WORKDIR /src/frameworks/elisabeth
RUN cargo build --locked --release --bench benchmark-elisabeth \
 && mkdir -p /out/bin \
 && find target/release -type f -name 'benchmark_elisabeth-*' ! -name '*.d' ! -name '*.pdb' -exec cp {} /out/bin/benchmark-elisabeth \; \
 && test -x /out/bin/benchmark-elisabeth \
 && strip /out/bin/benchmark-elisabeth \
 && install -D -m 0555 /usr/lib/x86_64-linux-gnu/libfftw3.so.3 /out/lib/libfftw3.so.3 \
 && install -D -m 0555 /usr/lib/x86_64-linux-gnu/libfftw3f.so.3 /out/lib/libfftw3f.so.3

FROM ${TOOLING_IMAGE} AS runtime
ARG IMAGE_REVISION=unknown
ARG IMAGE_VERSION=development
ARG TOOLING_DIGEST=development
LABEL org.opencontainers.image.title="HHE-ARENA Elisabeth HHE benchmark" \
      org.opencontainers.image.description="Prebuilt Elisabeth Rust homomorphic benchmark with the HHE-ARENA contract" \
      org.opencontainers.image.source="https://github.com/MarbleHE/HHE-ARENA" \
      org.opencontainers.image.revision="${IMAGE_REVISION}" \
      org.opencontainers.image.version="${IMAGE_VERSION}" \
      io.hhe-arena.contract.version="1" \
      io.hhe-arena.matrix.id="elisabeth-hhe-rust" \
      io.hhe-arena.manifest.path="/opt/hhe/benchmarks/manifest.json" \
      io.hhe-arena.results.path="/work/results" \
      io.hhe-arena.resource.class="heavy" \
      io.hhe-arena.tooling.digest="${TOOLING_DIGEST}"
USER root
COPY --from=builder /out/bin/benchmark-elisabeth /opt/hhe/frameworks/elisabeth-hhe/benchmark-elisabeth
COPY --from=builder /out/lib/ /opt/hhe/lib/
COPY manifests/elisabeth-hhe.v2.json /opt/hhe/benchmarks/manifest.json
ENV LD_LIBRARY_PATH=/opt/hhe/lib
RUN mkdir -p /work/results \
 && chown -R 65532:65532 /work \
 && hhe-validate-framework-image --require-benchmarks \
 && hhe list >/dev/null
USER 65532:65532
WORKDIR /work
ENTRYPOINT ["hhe-tooling"]
CMD ["--help"]
