# syntax=docker/dockerfile:1.7

ARG GCC_IMAGE=gcc:14.2-bookworm@sha256:b99b86a28812b1e6453a231a947dc43d76fe192788a12f344a9b568bf9f5d24c
ARG TOOLING_IMAGE=hhe-arena:tooling

FROM ${GCC_IMAGE} AS builder
RUN apt-get update \
 && apt-get install -y --no-install-recommends libfftw3-dev libgmp-dev libntl-dev make \
 && rm -rf /var/lib/apt/lists/*
WORKDIR /src
COPY frameworks/filip/sorting_hat/ frameworks/filip/sorting_hat/
WORKDIR /src/frameworks/filip/sorting_hat/src
RUN mkdir -p run \
 && make libfinal.a \
 && make test_homomorphic_filip \
 && install -D -m 0555 run/test_homomorphic_filip /out/test_homomorphic_filip \
 && install -D -m 0555 "$(gcc -print-file-name=libstdc++.so.6)" /out/lib/libstdc++.so.6 \
 && install -D -m 0555 /usr/lib/x86_64-linux-gnu/libntl.so.44 /out/lib/libntl.so.44 \
 && install -D -m 0555 /usr/lib/x86_64-linux-gnu/libgf2x.so.3 /out/lib/libgf2x.so.3 \
 && install -D -m 0555 /usr/lib/x86_64-linux-gnu/libfftw3.so.3 /out/lib/libfftw3.so.3

FROM ${TOOLING_IMAGE} AS runtime
ARG IMAGE_REVISION=unknown
ARG IMAGE_VERSION=development
ARG TOOLING_DIGEST=development
LABEL org.opencontainers.image.title="HHE-ARENA FiLIP Sorting Hat benchmarks" \
      org.opencontainers.image.description="Prebuilt FiLIP Sorting Hat C++ benchmarks with the HHE-ARENA contract" \
      org.opencontainers.image.source="https://github.com/MarbleHE/HHE-ARENA" \
      org.opencontainers.image.revision="${IMAGE_REVISION}" \
      org.opencontainers.image.version="${IMAGE_VERSION}" \
      io.hhe-arena.contract.version="1" \
      io.hhe-arena.matrix.id="filip-sorting-hat-cpp" \
      io.hhe-arena.manifest.path="/opt/hhe/benchmarks/manifest.json" \
      io.hhe-arena.results.path="/work/results" \
      io.hhe-arena.resource.class="extreme" \
      io.hhe-arena.tooling.digest="${TOOLING_DIGEST}"
USER root
COPY --from=builder /out/test_homomorphic_filip /opt/hhe/frameworks/filip-sorting-hat/test_homomorphic_filip
COPY --from=builder /out/lib/ /opt/hhe/lib/
COPY manifests/filip-sorting-hat.v2.json /opt/hhe/benchmarks/manifest.json
ENV LD_LIBRARY_PATH=/opt/hhe/lib
RUN mkdir -p /work/results \
 && chown -R 65532:65532 /work \
 && hhe-validate-framework-image --require-benchmarks \
 && hhe list >/dev/null
USER 65532:65532
WORKDIR /work
ENTRYPOINT ["hhe-tooling"]
CMD ["--help"]
