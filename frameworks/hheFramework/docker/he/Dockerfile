# syntax=docker/dockerfile:1.7

ARG GCC_IMAGE=gcc:14.2-bookworm@sha256:b99b86a28812b1e6453a231a947dc43d76fe192788a12f344a9b568bf9f5d24c
ARG TOOLING_IMAGE=hhe-arena:tooling

FROM ${GCC_IMAGE} AS builder
ARG DEBIAN_FRONTEND=noninteractive
ARG HELIB_REV=f0e3e010009c592cd411ba96baa8376eb485247a
ARG SEAL_REV=265d77e84fb316cc9ed88cda895e8964e802e854
ARG TFHE_REV=9373b5d8a5b022ca5b4e112ae1c2440bc18e160c
RUN apt-get update \
 && apt-get install -y --no-install-recommends \
      cmake=3.25.1-1 \
      git \
      libgmp-dev=2:6.2.1+dfsg1-1.1 \
      libm4ri-dev=20200125-1+b1 \
      libntl-dev=11.5.1-1+b2 \
 && rm -rf /var/lib/apt/lists/*
WORKDIR /deps
RUN git init HElib \
 && git -C HElib fetch --depth 1 https://github.com/homenc/HElib.git "${HELIB_REV}" \
 && git -C HElib checkout --detach FETCH_HEAD \
 && test "$(git -C HElib rev-parse HEAD)" = "${HELIB_REV}" \
 && cmake -S HElib -B HElib/build -DCMAKE_BUILD_TYPE=Release \
      -DCMAKE_INSTALL_PREFIX=/opt/helib -DBUILD_SHARED=ON \
      -DENABLE_THREADS=ON -DBUILD_TESTING=OFF \
 && cmake --build HElib/build --parallel 4 \
 && cmake --install HElib/build
RUN git init SEAL \
 && git -C SEAL fetch --depth 1 https://github.com/microsoft/SEAL.git "${SEAL_REV}" \
 && git -C SEAL checkout --detach FETCH_HEAD \
 && test "$(git -C SEAL rev-parse HEAD)" = "${SEAL_REV}" \
 && cmake -S SEAL -B SEAL/build -DCMAKE_BUILD_TYPE=Release \
      -DCMAKE_INSTALL_PREFIX=/opt/seal -DSEAL_BUILD_DEPS=OFF \
      -DSEAL_USE_MSGSL=OFF -DSEAL_USE_ZLIB=OFF -DSEAL_USE_ZSTD=OFF \
      -DSEAL_BUILD_EXAMPLES=OFF -DSEAL_BUILD_TESTS=OFF \
 && cmake --build SEAL/build --parallel 4 \
 && cmake --install SEAL/build
RUN git init tfhe \
 && git -C tfhe fetch --depth 1 https://github.com/tfhe/tfhe.git "${TFHE_REV}" \
 && git -C tfhe checkout --detach FETCH_HEAD \
 && test "$(git -C tfhe rev-parse HEAD)" = "${TFHE_REV}" \
 && cmake -S tfhe/src -B tfhe/build -DCMAKE_BUILD_TYPE=Release \
      -DCMAKE_INSTALL_PREFIX=/opt/tfhe -DENABLE_SPQLIOS_AVX=ON \
      -DENABLE_SPQLIOS_FMA=ON \
 && cmake --build tfhe/build --parallel 4 \
 && cmake --install tfhe/build

WORKDIR /src
COPY frameworks/hheFramework/ frameworks/hheFramework/
COPY benchmarking-utilities/c/utils/ benchmarking-utilities/c/utils/
RUN cmake -S frameworks/hheFramework -B /build \
      -DCMAKE_BUILD_TYPE=Release \
      -DBUILD_TESTING=OFF \
      -DHHE_FETCH_THIRDPARTY=OFF \
      -DHHE_BENCHMARK_UTILS_DIR=/src/benchmarking-utilities/c/utils \
      -DCMAKE_PREFIX_PATH='/opt/helib;/opt/seal' \
      -DTFHE_PREFIX=/opt/tfhe \
      -DM4RI_INCLUDE_DIR=/usr/include \
      -DM4RI_LIB=/usr/lib/x86_64-linux-gnu/libm4ri.so \
 && cmake --build /build --parallel 4
RUN mkdir -p /out/bin /out/lib \
 && find /build/tests -maxdepth 1 -type f \( -name '*_helib_test' -o -name '*_seal_test' -o -name '*_tfhe_test' \) \
      -exec install -m 0555 -t /out/bin {} + \
 && test "$(find /out/bin -maxdepth 1 -type f | wc -l)" -eq 55 \
 && find /opt/helib /opt/tfhe -type f -name '*.so*' -exec cp -P -t /out/lib {} + \
 && install -m 0555 /usr/lib/x86_64-linux-gnu/libm4ri-0.0.20200125.so /out/lib/ \
 && install -m 0555 /usr/lib/x86_64-linux-gnu/libntl.so.44 /out/lib/ \
 && install -m 0555 /usr/lib/x86_64-linux-gnu/libgf2x.so.3 /out/lib/ \
 && install -m 0555 /usr/lib/x86_64-linux-gnu/libgmp.so.10 /out/lib/ \
 && install -m 0555 "$(g++ -print-file-name=libstdc++.so.6)" /out/lib/libstdc++.so.6 \
 && install -m 0555 /usr/lib/x86_64-linux-gnu/libpng16.so.16 /out/lib/ \
 && install -m 0555 /lib/x86_64-linux-gnu/libz.so.1 /out/lib/

FROM ${TOOLING_IMAGE} AS runtime
ARG IMAGE_REVISION=unknown
ARG IMAGE_VERSION=development
ARG TOOLING_DIGEST=development
LABEL org.opencontainers.image.title="HHE-ARENA HHE Framework HE benchmarks" \
      org.opencontainers.image.source="https://github.com/MarbleHE/HHE-ARENA" \
      org.opencontainers.image.revision="${IMAGE_REVISION}" \
      org.opencontainers.image.version="${IMAGE_VERSION}" \
      io.hhe-arena.contract.version="1" \
      io.hhe-arena.matrix.id="hhe-framework-cpp" \
      io.hhe-arena.manifest.path="/opt/hhe/benchmarks/manifest.json" \
      io.hhe-arena.results.path="/work/results" \
      io.hhe-arena.resource.class="extreme" \
      io.hhe-arena.tooling.digest="${TOOLING_DIGEST}"
USER root
COPY --from=builder /out/bin/ /opt/hhe/frameworks/hhe-framework/
COPY --from=builder /out/lib/ /opt/hhe/lib/
COPY manifests/hhe-framework-he.v2.json /opt/hhe/benchmarks/manifest.json
COPY schemas/benchmark-manifest.schema.json /opt/hhe/schemas/benchmark-manifest.schema.json
COPY docker/validate_framework_image.py /opt/hhe/bin/hhe-validate-framework-image
ENV LD_LIBRARY_PATH=/opt/hhe/lib
RUN chmod 0555 /opt/hhe/bin/hhe-validate-framework-image \
 && mkdir -p /work/results \
 && chown -R 65532:65532 /work \
 && test "$(find /opt/hhe/frameworks/hhe-framework -maxdepth 1 -type f | wc -l)" -eq 55 \
 && for executable in /opt/hhe/frameworks/hhe-framework/*; do \
      if ldd "$executable" 2>&1 | grep -q 'not found'; then \
        echo "missing runtime library: $executable"; exit 1; \
      fi; \
    done \
 && hhe-validate-framework-image --require-benchmarks \
 && hhe list >/dev/null
USER 65532:65532
WORKDIR /work
ENTRYPOINT ["hhe-tooling"]
CMD ["--help"]
