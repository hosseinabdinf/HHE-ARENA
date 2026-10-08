# syntax=docker/dockerfile:1.7

ARG GCC_IMAGE=gcc:14.2-bookworm@sha256:b99b86a28812b1e6453a231a947dc43d76fe192788a12f344a9b568bf9f5d24c
ARG TOOLING_IMAGE=hhe-arena:tooling

FROM ${GCC_IMAGE} AS builder
ARG DEBIAN_FRONTEND=noninteractive
ARG HELIB_REV=f0e3e010009c592cd411ba96baa8376eb485247a
ARG TFHE_REV=9373b5d8a5b022ca5b4e112ae1c2440bc18e160c
RUN apt-get update \
 && apt-get install -y --no-install-recommends \
      cmake=3.25.1-1 \
      git \
      libgf2x3=1.3.0-2 \
      libgmp-dev=2:6.2.1+dfsg1-1.1 \
      libm4ri-dev=20200125-1+b1 \
      libntl-dev=11.5.1-1+b2 \
      libssl-dev=3.0.22-1~deb12u1 \
      pkg-config \
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
COPY frameworks/fasta/ frameworks/fasta/
RUN cmake -S frameworks/fasta -B /src/build -DCMAKE_BUILD_TYPE=Release \
      -DFASTA_FETCH_THIRDPARTY=OFF \
      -DCMAKE_PREFIX_PATH=/opt/helib \
      -DTFHE_PREFIX=/opt/tfhe \
      -DM4RI_INCLUDE_DIR=/usr/include \
      -DM4RI_LIB=/usr/lib/x86_64-linux-gnu/libm4ri.so \
 && cmake --build /src/build --parallel 4
RUN mkdir -p /out/bin /out/lib \
 && install -D -m 0555 /src/build/fasta_helib_bin/fasta_helib_test /out/bin/fasta-helib \
 && install -D -m 0555 /src/build/rasta_helib_bin/rasta_helib_test /out/bin/rasta-helib \
 && install -D -m 0555 /src/build/rasta_pack_helib_bin/rasta_pack_helib_test /out/bin/rasta-pack-helib \
 && install -D -m 0555 /src/build/rasta_tfhe_bin/rasta_tfhe_test /out/bin/rasta-tfhe \
 && find /opt/helib /opt/tfhe -type f -name '*.so*' -exec cp -P -t /out/lib {} + \
 && install -m 0555 /usr/lib/x86_64-linux-gnu/libm4ri-0.0.20200125.so /out/lib/ \
 && install -m 0555 /usr/lib/x86_64-linux-gnu/libntl.so.44 /out/lib/ \
 && install -m 0555 /usr/lib/x86_64-linux-gnu/libgf2x.so.3 /out/lib/ \
 && install -m 0555 /usr/lib/x86_64-linux-gnu/libgmp.so.10 /out/lib/ \
 && install -m 0555 /usr/lib/x86_64-linux-gnu/libpng16.so.16 /out/lib/ \
 && install -m 0555 /lib/x86_64-linux-gnu/libz.so.1 /out/lib/ \
 && install -m 0555 "$(g++ -print-file-name=libstdc++.so.6)" /out/lib/libstdc++.so.6

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
COPY --from=builder /out/bin/fasta-helib /opt/hhe/frameworks/fasta-rasta/fasta-helib
COPY --from=builder /out/bin/rasta-helib /opt/hhe/frameworks/fasta-rasta/rasta-helib
COPY --from=builder /out/bin/rasta-pack-helib /opt/hhe/frameworks/fasta-rasta/rasta-pack-helib
COPY --from=builder /out/bin/rasta-tfhe /opt/hhe/frameworks/fasta-rasta/rasta-tfhe-bin
COPY --from=builder /out/lib/ /opt/hhe/lib/
ENV LD_LIBRARY_PATH=/opt/hhe/lib
COPY frameworks/fasta/hhe/rasta/matrix_*.txt /opt/hhe/frameworks/fasta-rasta/rasta/
COPY frameworks/fasta/hhe/rasta_pack/matrix_*.txt /opt/hhe/frameworks/fasta-rasta/rasta-pack/
COPY frameworks/fasta/hhe/rasta_tfhe/matrix_*.txt /opt/hhe/frameworks/fasta-rasta/rasta-tfhe/
COPY manifests/fasta-rasta.v2.json /opt/hhe/benchmarks/manifest.json
RUN mkdir -p /work/results \
 && chown -R 65532:65532 /work \
 && hhe-validate-framework-image --require-benchmarks --check-links \
 && hhe list >/dev/null
USER 65532:65532
WORKDIR /work
ENTRYPOINT ["hhe-tooling"]
CMD ["--help"]
