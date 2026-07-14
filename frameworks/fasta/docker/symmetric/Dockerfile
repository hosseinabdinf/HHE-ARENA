# syntax=docker/dockerfile:1.7

ARG GCC_IMAGE=gcc:14.2-bookworm@sha256:b99b86a28812b1e6453a231a947dc43d76fe192788a12f344a9b568bf9f5d24c
ARG TOOLING_IMAGE=hhe-arena:tooling

FROM ${GCC_IMAGE} AS builder
WORKDIR /src
COPY frameworks/fasta/sym/ ./sym/
COPY frameworks/fasta/utils/ ./utils/
RUN mkdir -p /out/lib \
 && gcc -std=c11 -O3 -DNDEBUG -D_DEFAULT_SOURCE -Wall -Wextra -Werror \
      -Isym -Iutils \
      sym/fasta_sym.c sym/Fasta.c sym/bitVector.c utils/utils.c \
      -lgmp -o /out/fasta-symmetric \
 && install -D -m 0555 /usr/lib/x86_64-linux-gnu/libgmp.so.10 /out/lib/libgmp.so.10

FROM ${TOOLING_IMAGE} AS runtime
ARG IMAGE_REVISION=unknown
ARG IMAGE_VERSION=development
ARG TOOLING_DIGEST=development
LABEL org.opencontainers.image.title="HHE-ARENA Fasta symmetric benchmark" \
      org.opencontainers.image.description="Prebuilt Fasta native C encrypt/decrypt benchmark with the HHE-ARENA contract" \
      org.opencontainers.image.source="https://github.com/MarbleHE/HHE-ARENA" \
      org.opencontainers.image.revision="${IMAGE_REVISION}" \
      org.opencontainers.image.version="${IMAGE_VERSION}" \
      io.hhe-arena.contract.version="1" \
      io.hhe-arena.matrix.id="fasta-symmetric-c" \
      io.hhe-arena.manifest.path="/opt/hhe/benchmarks/manifest.json" \
      io.hhe-arena.results.path="/work/results" \
      io.hhe-arena.resource.class="light" \
      io.hhe-arena.tooling.digest="${TOOLING_DIGEST}"
USER root
COPY --from=builder /out/fasta-symmetric /opt/hhe/frameworks/fasta-symmetric/fasta-symmetric
COPY --from=builder /out/lib/ /opt/hhe/lib/
COPY frameworks/fasta/sym/originalPlaintext.txt /opt/hhe/frameworks/fasta-symmetric/originalPlaintext.txt
COPY manifests/fasta-symmetric.v2.json /opt/hhe/benchmarks/manifest.json
COPY schemas/benchmark-manifest.schema.json /opt/hhe/schemas/benchmark-manifest.schema.json
COPY docker/validate_framework_image.py /opt/hhe/bin/hhe-validate-framework-image
ENV LD_LIBRARY_PATH=/opt/hhe/lib
RUN chmod 0555 /opt/hhe/bin/hhe-validate-framework-image \
 && mkdir -p /work/results \
 && chown -R 65532:65532 /work \
 && ldd /opt/hhe/frameworks/fasta-symmetric/fasta-symmetric | grep -q '/opt/hhe/lib/libgmp.so.10' \
 && hhe-validate-framework-image --require-benchmarks \
 && hhe list >/dev/null
USER 65532:65532
WORKDIR /work
ENTRYPOINT ["hhe-tooling"]
CMD ["--help"]
