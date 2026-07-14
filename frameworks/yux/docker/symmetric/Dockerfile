# syntax=docker/dockerfile:1.7

ARG GCC_IMAGE=gcc:14.2-bookworm@sha256:b99b86a28812b1e6453a231a947dc43d76fe192788a12f344a9b568bf9f5d24c
ARG TOOLING_IMAGE=hhe-arena:tooling

FROM ${GCC_IMAGE} AS builder
ARG DEBIAN_FRONTEND=noninteractive
ARG CMAKE_VERSION=3.25.1-1
ARG NTL_VERSION=11.5.1-1+b2
RUN apt-get update \
 && apt-get install -y --no-install-recommends \
      cmake=${CMAKE_VERSION} \
      libntl-dev=${NTL_VERSION} \
 && rm -rf /var/lib/apt/lists/*
WORKDIR /src
COPY frameworks/yux/Yux/ frameworks/yux/Yux/
COPY frameworks/yux/tests/test-Yu2x-8.cpp frameworks/yux/tests/test-Yu2x-16.cpp frameworks/yux/tests/test-Yupx-p.cpp frameworks/yux/tests/
COPY frameworks/yux/utils/ frameworks/yux/utils/
COPY frameworks/yux/docker/symmetric/CMakeLists.txt frameworks/yux/docker/symmetric/CMakeLists.txt
RUN cmake -S frameworks/yux/docker/symmetric -B /build \
      -DCMAKE_BUILD_TYPE=Release \
      -DCMAKE_C_FLAGS_RELEASE="-O3 -DNDEBUG" \
      -DCMAKE_CXX_FLAGS_RELEASE="-O3 -DNDEBUG" \
      -DYUX_SOURCE=/src/frameworks/yux \
 && cmake --build /build --parallel 2 \
 && cmake --install /build --prefix /out \
 && install -D -m 0555 /usr/lib/x86_64-linux-gnu/libntl.so.44 /out/lib/libntl.so.44 \
 && install -D -m 0555 /usr/lib/x86_64-linux-gnu/libgf2x.so.3 /out/lib/libgf2x.so.3 \
 && install -D -m 0555 /lib/x86_64-linux-gnu/libgmp.so.10 /out/lib/libgmp.so.10 \
 && install -D -m 0555 /usr/local/lib64/libstdc++.so.6 /out/lib/libstdc++.so.6 \
 && install -D -m 0555 /usr/local/lib64/libgcc_s.so.1 /out/lib/libgcc_s.so.1

FROM ${TOOLING_IMAGE} AS runtime
ARG IMAGE_REVISION=unknown
ARG IMAGE_VERSION=development
ARG TOOLING_DIGEST=development
LABEL org.opencontainers.image.title="HHE-ARENA YuX symmetric benchmarks" \
      org.opencontainers.image.description="Prebuilt YuX native C++ symmetric variants with the HHE-ARENA contract" \
      org.opencontainers.image.source="https://github.com/MarbleHE/HHE-ARENA" \
      org.opencontainers.image.revision="${IMAGE_REVISION}" \
      org.opencontainers.image.version="${IMAGE_VERSION}" \
      io.hhe-arena.contract.version="1" \
      io.hhe-arena.matrix.id="yux-symmetric-cpp" \
      io.hhe-arena.manifest.path="/opt/hhe/benchmarks/manifest.json" \
      io.hhe-arena.results.path="/work/results" \
      io.hhe-arena.resource.class="light" \
      io.hhe-arena.tooling.digest="${TOOLING_DIGEST}"
USER root
COPY --from=builder /out/bin/ /opt/hhe/frameworks/yux-symmetric/
COPY --from=builder /out/lib/ /opt/hhe/lib/
COPY manifests/yux-symmetric.v2.json /opt/hhe/benchmarks/manifest.json
COPY schemas/benchmark-manifest.schema.json /opt/hhe/schemas/benchmark-manifest.schema.json
COPY docker/validate_framework_image.py /opt/hhe/bin/hhe-validate-framework-image
ENV LD_LIBRARY_PATH=/opt/hhe/lib
RUN chmod 0555 /opt/hhe/bin/hhe-validate-framework-image \
 && mkdir -p /work/results \
 && chown -R 65532:65532 /work \
 && test "$(find /opt/hhe/frameworks/yux-symmetric -maxdepth 1 -type f | wc -l)" -eq 3 \
 && for executable in /opt/hhe/frameworks/yux-symmetric/*; do \
      ldd "$executable" | grep -q '/opt/hhe/lib/libstdc++.so.6'; \
      ! ldd "$executable" 2>&1 | grep -Eq 'not found|version .* not found'; \
    done \
 && hhe-validate-framework-image --require-benchmarks \
 && hhe list >/dev/null
USER 65532:65532
WORKDIR /work
ENTRYPOINT ["hhe-tooling"]
CMD ["--help"]
