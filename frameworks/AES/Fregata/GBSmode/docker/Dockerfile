# syntax=docker/dockerfile:1.7

ARG GCC_IMAGE=gcc:14.2-bookworm@sha256:b99b86a28812b1e6453a231a947dc43d76fe192788a12f344a9b568bf9f5d24c
ARG TOOLING_IMAGE=hhe-arena:tooling

FROM ${GCC_IMAGE} AS builder
ARG DEBIAN_FRONTEND=noninteractive
ARG CMAKE_VERSION=3.25.1-1
ARG GIT_VERSION=1:2.39.5-0+deb12u2
ARG TFHE_COMMIT=ad518fe17f1b1578d0c603eb16d51cb348a4406e
RUN apt-get update \
 && apt-get install -y --no-install-recommends cmake=${CMAKE_VERSION} git=${GIT_VERSION} \
 && rm -rf /var/lib/apt/lists/*
WORKDIR /src
RUN git init /tmp/tfhe \
 && git -C /tmp/tfhe remote add origin https://github.com/tfhe/tfhe.git \
 && git -C /tmp/tfhe fetch --depth 1 origin "${TFHE_COMMIT}" \
 && git -C /tmp/tfhe checkout --detach FETCH_HEAD \
 && test "$(git -C /tmp/tfhe rev-parse HEAD)" = "${TFHE_COMMIT}" \
 && cmake -S /tmp/tfhe/src -B /tmp/tfhe/build \
      -DCMAKE_BUILD_TYPE=Release \
      -DCMAKE_CXX_FLAGS=-Wno-error=restrict \
      -DENABLE_SPQLIOS_AVX=on \
      -DENABLE_SPQLIOS_FMA=on \
      -DENABLE_TESTS=off \
      -DCMAKE_INSTALL_PREFIX=/opt/tfhe \
 && cmake --build /tmp/tfhe/build --parallel 2 \
 && cmake --install /tmp/tfhe/build
COPY benchmarking-utilities/c/utils/ benchmarking-utilities/c/utils/
COPY benchmarking-utilities/c/tests/schema_smoke.c benchmarking-utilities/c/tests/schema_smoke.c
COPY frameworks/AES/Fregata/stateful_benchmark.h frameworks/AES/Fregata/stateful_benchmark.h
COPY frameworks/AES/Fregata/GBSmode/ frameworks/AES/Fregata/GBSmode/
RUN cmake -S frameworks/AES/Fregata/GBSmode -B /build \
      -DCMAKE_BUILD_TYPE=Release \
      -DCMAKE_CXX_FLAGS=-I/opt/tfhe/include \
      -DCMAKE_EXE_LINKER_FLAGS=-L/opt/tfhe/lib \
 && cmake --build /build --target homoAES --parallel 2 \
 && install -D -m 0555 /build/homoAES /out/bin/homoAES \
 && strip /out/bin/homoAES \
 && install -D -m 0555 /opt/tfhe/lib/libtfhe-spqlios-fma.so /out/lib/libtfhe-spqlios-fma.so \
 && install -D -m 0555 /usr/local/lib64/libstdc++.so.6 /out/lib/libstdc++.so.6 \
 && install -D -m 0555 /usr/local/lib64/libgcc_s.so.1 /out/lib/libgcc_s.so.1 \
 && install -D -m 0555 /usr/local/lib64/libgomp.so.1 /out/lib/libgomp.so.1

FROM ${TOOLING_IMAGE} AS runtime
ARG IMAGE_REVISION=unknown
ARG IMAGE_VERSION=development
ARG TOOLING_DIGEST=development
LABEL org.opencontainers.image.title="HHE-ARENA Fregata GBS AES benchmark" \
      org.opencontainers.image.description="Prebuilt Fregata AES gate-bootstrapping benchmark with the HHE-ARENA contract" \
      org.opencontainers.image.source="https://github.com/MarbleHE/HHE-ARENA" \
      org.opencontainers.image.revision="${IMAGE_REVISION}" \
      org.opencontainers.image.version="${IMAGE_VERSION}" \
      io.hhe-arena.contract.version="1" \
      io.hhe-arena.matrix.id="fregata-gbs-aes-tfhepp-cpp" \
      io.hhe-arena.manifest.path="/opt/hhe/benchmarks/manifest.json" \
      io.hhe-arena.results.path="/work/results" \
      io.hhe-arena.resource.class="extreme" \
      io.hhe-arena.tooling.digest="${TOOLING_DIGEST}"
USER root
COPY --from=builder /out/bin/homoAES /opt/hhe/frameworks/fregata-gbs-aes/homoAES
COPY --from=builder /out/lib/ /opt/hhe/lib/
COPY manifests/fregata-gbs-aes.v2.json /opt/hhe/benchmarks/manifest.json
ENV LD_LIBRARY_PATH=/opt/hhe/lib
RUN mkdir -p /work/results \
 && chown -R 65532:65532 /work \
 && ldd /opt/hhe/frameworks/fregata-gbs-aes/homoAES | grep -q '/opt/hhe/lib/libtfhe-spqlios-fma.so' \
 && ! ldd /opt/hhe/frameworks/fregata-gbs-aes/homoAES 2>&1 | grep -Eq 'not found|version .* not found' \
 && hhe-validate-framework-image --require-benchmarks \
 && hhe list >/dev/null
USER 65532:65532
WORKDIR /work
ENTRYPOINT ["hhe-tooling"]
CMD ["--help"]
