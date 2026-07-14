# syntax=docker/dockerfile:1.7

ARG PYTHON_IMAGE=python:3.13-slim-bookworm@sha256:fcbd8dfc2605ba7c2eca646846c5e892b2931e41f6227985154a596f26ab8ed7
ARG TOOLING_IMAGE=hhe-arena:tooling

FROM ${PYTHON_IMAGE} AS counter-builder
ARG DEBIAN_FRONTEND=noninteractive
ARG GCC_VERSION=4:12.2.0-3
ARG LIBC6_DEV_VERSION=2.36-9+deb12u14
RUN apt-get update \
 && apt-get install -y --no-install-recommends \
      gcc=${GCC_VERSION} \
      libc6-dev=${LIBC6_DEV_VERSION} \
 && rm -rf /var/lib/apt/lists/* \
 && pip install --no-cache-dir cffi==2.1.0 setuptools==80.9.0
WORKDIR /app
COPY frameworks/margrethe/scripts/serialized_counter.py ./
RUN PYTHONPATH=/app python -c 'import serialized_counter; assert serialized_counter.available()'

FROM ${TOOLING_IMAGE} AS runtime
ARG IMAGE_REVISION=unknown
ARG IMAGE_VERSION=development
ARG TOOLING_DIGEST=development
LABEL org.opencontainers.image.title="HHE-ARENA Margrethe symmetric benchmark" \
      org.opencontainers.image.description="Prebuilt Margrethe Python SKE benchmark with the HHE-ARENA contract" \
      org.opencontainers.image.source="https://github.com/MarbleHE/HHE-ARENA" \
      org.opencontainers.image.revision="${IMAGE_REVISION}" \
      org.opencontainers.image.version="${IMAGE_VERSION}" \
      io.hhe-arena.contract.version="1" \
      io.hhe-arena.matrix.id="margrethe-ske-python" \
      io.hhe-arena.manifest.path="/opt/hhe/benchmarks/manifest.json" \
      io.hhe-arena.results.path="/work/results" \
      io.hhe-arena.resource.class="light" \
      io.hhe-arena.tooling.digest="${TOOLING_DIGEST}"
USER root
COPY --from=counter-builder /app/ /opt/hhe/frameworks/margrethe-ske/
COPY frameworks/margrethe/scripts/cipher.py /opt/hhe/frameworks/margrethe-ske/cipher.py
COPY frameworks/margrethe/scripts/benchmark_utils.py /opt/hhe/frameworks/margrethe-ske/benchmark_utils.py
COPY frameworks/margrethe/scripts/run /opt/hhe/frameworks/margrethe-ske/run
COPY manifests/margrethe-symmetric.v2.json /opt/hhe/benchmarks/manifest.json
RUN chmod 0555 /opt/hhe/frameworks/margrethe-ske/run \
 && mkdir -p /work/results \
 && chown -R 65532:65532 /work \
 && PYTHONPATH=/opt/hhe/frameworks/margrethe-ske python -c 'import serialized_counter; assert serialized_counter.available()' \
 && PYTHONPATH=/opt/hhe/frameworks/margrethe-ske python -c 'import cipher; assert cipher.Cipher.__name__ == "Cipher"' \
 && hhe-validate-framework-image --require-benchmarks \
 && hhe list >/dev/null
USER 65532:65532
WORKDIR /work
ENTRYPOINT ["hhe-tooling"]
CMD ["--help"]
