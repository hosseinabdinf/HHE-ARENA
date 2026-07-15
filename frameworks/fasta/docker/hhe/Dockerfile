# syntax=docker/dockerfile:1.7

ARG GCC_IMAGE=gcc:14.2-bookworm@sha256:b99b86a28812b1e6453a231a947dc43d76fe192788a12f344a9b568bf9f5d24c
ARG TOOLING_IMAGE=hhe-arena:tooling

FROM ${GCC_IMAGE} AS builder
ARG DEBIAN_FRONTEND=noninteractive
RUN apt-get update \
 && apt-get install -y --no-install-recommends \
      autoconf automake cmake git libssl-dev libtool make pkg-config wget \
 && rm -rf /var/lib/apt/lists/*
WORKDIR /src
COPY frameworks/fasta/ frameworks/fasta/
RUN sed -i 's/[[:cntrl:]]$//' frameworks/fasta/thirdparty/build.sh && cd frameworks/fasta && bash thirdparty/build.sh
RUN cmake -S frameworks/fasta -B /src/build -DCMAKE_BUILD_TYPE=Release \
 && cmake --build /src/build --parallel \
 && install -D -m 0555 /src/build/fasta_helib_bin/fasta_helib_test /out/fasta-helib \
 && install -D -m 0555 /src/build/rasta_helib_bin/rasta_helib_test /out/rasta-helib \
 && install -D -m 0555 /src/build/rasta_pack_helib_bin/rasta_pack_helib_test /out/rasta-pack-helib \
 && install -D -m 0555 /src/build/rasta_tfhe_bin/rasta_tfhe_test /out/rasta-tfhe

FROM ${TOOLING_IMAGE} AS runtime
ARG IMAGE_REVISION=unknown
ARG IMAGE_VERSION=development
ARG TOOLING_DIGEST=development
LABEL org.opencontainers.image.title="HHE-ARENA Fasta/Rasta HElib and TFHE benchmarks" \
      org.opencontainers.image.description="Prebuilt Fasta/Rasta homomorphic benchmarks with the HHE-ARENA contract" \
      org.opencontainers.image.source="https://github.com/MarbleHE/HHE-ARENA" \
      org.opencontainers.image.revision="${IMAGE_REVISION}" \
      org.opencontainers.image.version="${IMAGE_VERSION}" \
      io.hhe-arena.contract.version="1" \
      io.hhe-arena.matrix.id="fasta-rasta-c-cpp" \
      io.hhe-arena.manifest.path="/opt/hhe/benchmarks/manifest.json" \
      io.hhe-arena.results.path="/work/results" \
      io.hhe-arena.resource.class="extreme" \
      io.hhe-arena.tooling.digest="${TOOLING_DIGEST}"
USER root
COPY --from=builder /out/fasta-helib /opt/hhe/frameworks/fasta-rasta/fasta-helib
COPY --from=builder /out/rasta-helib /opt/hhe/frameworks/fasta-rasta/rasta-helib
COPY --from=builder /out/rasta-pack-helib /opt/hhe/frameworks/fasta-rasta/rasta-pack-helib
COPY --from=builder /out/rasta-tfhe /opt/hhe/frameworks/fasta-rasta/rasta-tfhe-bin
COPY --from=builder /src/frameworks/fasta/thirdparty/HElib/install/lib/ /opt/hhe/lib/
COPY --from=builder /src/frameworks/fasta/thirdparty/NTL/usr/local/lib/ /opt/hhe/lib/
COPY --from=builder /src/frameworks/fasta/thirdparty/GMP/usr/local/lib/ /opt/hhe/lib/
COPY --from=builder /src/frameworks/fasta/thirdparty/m4ri/installed/lib/ /opt/hhe/lib/
COPY --from=builder /src/frameworks/fasta/thirdparty/tfhe/installed/lib/ /opt/hhe/lib/
ENV LD_LIBRARY_PATH=/opt/hhe/lib
COPY frameworks/fasta/hhe/rasta/matrix_*.txt /opt/hhe/frameworks/fasta-rasta/rasta/
COPY frameworks/fasta/hhe/rasta_pack/matrix_*.txt /opt/hhe/frameworks/fasta-rasta/rasta-pack/
COPY frameworks/fasta/hhe/rasta_tfhe/matrix_*.txt /opt/hhe/frameworks/fasta-rasta/rasta-tfhe/
COPY manifests/fasta-rasta.v2.json /opt/hhe/benchmarks/manifest.json
RUN mkdir -p /work/results \
 && chown -R 65532:65532 /work \
 && hhe-validate-framework-image --require-benchmarks \
 && hhe list >/dev/null
USER 65532:65532
WORKDIR /work
ENTRYPOINT ["hhe-tooling"]
CMD ["--help"]
