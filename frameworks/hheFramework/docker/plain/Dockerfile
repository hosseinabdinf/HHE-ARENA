# syntax=docker/dockerfile:1.7

ARG GCC_IMAGE=gcc:14.2-bookworm@sha256:b99b86a28812b1e6453a231a947dc43d76fe192788a12f344a9b568bf9f5d24c
ARG TOOLING_IMAGE=hhe-arena:tooling

FROM ${GCC_IMAGE} AS builder
ARG DEBIAN_FRONTEND=noninteractive
ARG CMAKE_VERSION=3.25.1-1
ARG M4RI_VERSION=20200125-1+b1
RUN apt-get update \
 && apt-get install -y --no-install-recommends \
      cmake=${CMAKE_VERSION} \
      libm4ri-dev=${M4RI_VERSION} \
 && rm -rf /var/lib/apt/lists/*
WORKDIR /src
COPY frameworks/hheFramework/ frameworks/hheFramework/
COPY benchmarking-utilities/c/utils/ benchmarking-utilities/c/utils/
RUN cmake -S frameworks/hheFramework/docker/plain -B /build \
      -DCMAKE_BUILD_TYPE=Release \
      -DCMAKE_C_FLAGS_RELEASE="-O3 -DNDEBUG" \
      -DCMAKE_CXX_FLAGS_RELEASE="-O3 -DNDEBUG" \
      -DHHE_FRAMEWORK_SOURCE=/src/frameworks/hheFramework \
      -DHHE_BENCHMARK_UTILS_DIR=/src/benchmarking-utilities/c/utils \
 && cmake --build /build --parallel 2 \
 && cmake --install /build --prefix /out \
 && install -D -m 0555 /usr/lib/x86_64-linux-gnu/libm4ri-0.0.20200125.so /out/lib/libm4ri-0.0.20200125.so \
 && install -D -m 0555 "$(gcc -print-file-name=libstdc++.so.6)" /out/lib/libstdc++.so.6 \
 && install -D -m 0555 /lib/x86_64-linux-gnu/libpng16.so.16 /out/lib/libpng16.so.16 \
 && install -D -m 0555 /lib/x86_64-linux-gnu/libz.so.1 /out/lib/libz.so.1

FROM ${TOOLING_IMAGE} AS runtime
ARG IMAGE_REVISION=unknown
ARG IMAGE_VERSION=development
ARG TOOLING_DIGEST=development
LABEL org.opencontainers.image.title="HHE-ARENA HHE Framework plain benchmarks" \
      org.opencontainers.image.description="Prebuilt HHE Framework plain C++ cipher variants with the HHE-ARENA contract" \
      org.opencontainers.image.source="https://github.com/MarbleHE/HHE-ARENA" \
      org.opencontainers.image.revision="${IMAGE_REVISION}" \
      org.opencontainers.image.version="${IMAGE_VERSION}" \
      io.hhe-arena.contract.version="1" \
      io.hhe-arena.matrix.id="hhe-framework-plain-cpp" \
      io.hhe-arena.manifest.path="/opt/hhe/benchmarks/manifest.json" \
      io.hhe-arena.results.path="/work/results" \
      io.hhe-arena.resource.class="light" \
      io.hhe-arena.tooling.digest="${TOOLING_DIGEST}"
USER root
COPY --from=builder /out/bin/ /opt/hhe/frameworks/hhe-framework-plain/
COPY --from=builder /out/lib/ /opt/hhe/lib/
COPY manifests/hhe-framework-plain.v2.json /opt/hhe/benchmarks/manifest.json
COPY schemas/benchmark-manifest.schema.json /opt/hhe/schemas/benchmark-manifest.schema.json
COPY docker/validate_framework_image.py /opt/hhe/bin/hhe-validate-framework-image
ENV LD_LIBRARY_PATH=/opt/hhe/lib
RUN chmod 0555 /opt/hhe/bin/hhe-validate-framework-image \
 && mkdir -p /work/results \
 && chown -R 65532:65532 /work \
 && test "$(find /opt/hhe/frameworks/hhe-framework-plain -maxdepth 1 -type f | wc -l)" -eq 23 \
 && hhe-validate-framework-image --require-benchmarks \
 && hhe list >/dev/null
USER 65532:65532
WORKDIR /work
ENTRYPOINT ["hhe-tooling"]
CMD ["--help"]
