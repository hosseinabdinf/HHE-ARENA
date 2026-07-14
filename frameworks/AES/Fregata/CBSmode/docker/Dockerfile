# syntax=docker/dockerfile:1.7

ARG GCC_IMAGE=gcc:14.2-bookworm@sha256:b99b86a28812b1e6453a231a947dc43d76fe192788a12f344a9b568bf9f5d24c
ARG TOOLING_IMAGE=hhe-arena:tooling

FROM ${GCC_IMAGE} AS builder
ARG DEBIAN_FRONTEND=noninteractive
ARG CMAKE_VERSION=3.25.1-1
RUN apt-get update \
 && apt-get install -y --no-install-recommends cmake=${CMAKE_VERSION} \
 && rm -rf /var/lib/apt/lists/*
WORKDIR /src
COPY frameworks/AES/Fregata/CBSmode/ frameworks/AES/Fregata/CBSmode/
RUN cmake -S frameworks/AES/Fregata/CBSmode -B /build \
      -DCMAKE_BUILD_TYPE=Release \
      -DENABLE_TEST=OFF \
      -DENABLE_BENCHMARK=OFF \
      -DENABLE_TUTORIAL=OFF \
 && cmake --build /build --target sym_aes hom_aes --parallel 2 \
 && install -D -m 0555 /build/homoSM4_CB/sym_aes /out/bin/sym_aes \
 && install -D -m 0555 /build/homoSM4_CB/hom_aes /out/bin/hom_aes \
 && strip /out/bin/sym_aes /out/bin/hom_aes \
 && install -D -m 0555 /usr/local/lib64/libstdc++.so.6 /out/lib/libstdc++.so.6 \
 && install -D -m 0555 /usr/local/lib64/libgcc_s.so.1 /out/lib/libgcc_s.so.1 \
 && install -D -m 0555 /usr/local/lib64/libgomp.so.1 /out/lib/libgomp.so.1

FROM ${TOOLING_IMAGE} AS runtime
ARG IMAGE_REVISION=unknown
ARG IMAGE_VERSION=development
ARG TOOLING_DIGEST=development
LABEL org.opencontainers.image.title="HHE-ARENA Fregata CBS AES benchmark" \
      org.opencontainers.image.description="Prebuilt Fregata AES/TFHE++ circuit-bootstrapping benchmarks with the HHE-ARENA contract" \
      org.opencontainers.image.source="https://github.com/MarbleHE/HHE-ARENA" \
      org.opencontainers.image.revision="${IMAGE_REVISION}" \
      org.opencontainers.image.version="${IMAGE_VERSION}" \
      io.hhe-arena.contract.version="1" \
      io.hhe-arena.matrix.id="fregata-cbs-aes-tfhepp-cpp" \
      io.hhe-arena.manifest.path="/opt/hhe/benchmarks/manifest.json" \
      io.hhe-arena.results.path="/work/results" \
      io.hhe-arena.resource.class="extreme" \
      io.hhe-arena.tooling.digest="${TOOLING_DIGEST}"
USER root
COPY --from=builder /out/bin/ /opt/hhe/frameworks/fregata-cbs/
COPY --from=builder /out/lib/ /opt/hhe/lib/
COPY manifests/fregata-cbs-aes.v2.json /opt/hhe/benchmarks/manifest.json
ENV LD_LIBRARY_PATH=/opt/hhe/lib
RUN mkdir -p /work/results \
 && chown -R 65532:65532 /work \
 && ! ldd /opt/hhe/frameworks/fregata-cbs/sym_aes 2>&1 | grep -Eq 'not found|version .* not found' \
 && ! ldd /opt/hhe/frameworks/fregata-cbs/hom_aes 2>&1 | grep -Eq 'not found|version .* not found' \
 && hhe-validate-framework-image --require-benchmarks \
 && hhe list >/dev/null
USER 65532:65532
WORKDIR /work
ENTRYPOINT ["hhe-tooling"]
CMD ["--help"]
