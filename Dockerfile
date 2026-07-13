# syntax=docker/dockerfile:1.7

ARG GCC_IMAGE=gcc:14.2-bookworm@sha256:b99b86a28812b1e6453a231a947dc43d76fe192788a12f344a9b568bf9f5d24c
ARG GO_IMAGE=golang:1.22-bookworm@sha256:3d699e4d15d0f8f13c9195c0632a16702b8cbdece2955af1c23b37ae5d55a253
ARG RUST_IMAGE=rust:1.85-bookworm@sha256:e51d0265072d2d9d5d320f6a44dde6b9ef13653b035098febd68cce8fa7c0bc4
ARG PYTHON_IMAGE=python:3.13-slim-bookworm@sha256:fcbd8dfc2605ba7c2eca646846c5e892b2931e41f6227985154a596f26ab8ed7

FROM ${GCC_IMAGE} AS c-builder
WORKDIR /src
COPY benchmarking-utilities/c/utils/ ./utils/
COPY benchmarking-utilities/c/tests/schema_smoke.c ./tests/schema_smoke.c
RUN mkdir -p /out \
 && gcc -std=c11 -O2 -Wall -Wextra -Werror -Iutils -c utils/utils.c -o /tmp/utils.o \
 && g++ -std=c++17 -O2 -Wall -Wextra -Werror -Iutils -c utils/printer.cpp -o /tmp/printer.o \
 && ar rcs /out/libhhe_benchmark_utils.a /tmp/utils.o /tmp/printer.o \
 && g++ -std=c++17 -O2 -Wall -Wextra -Werror -Iutils tests/schema_smoke.c /out/libhhe_benchmark_utils.a -o /out/hhe-c-schema-smoke

FROM ${GO_IMAGE} AS go-builder
WORKDIR /src
COPY benchmarking-utilities/go/go.mod benchmarking-utilities/go/go.sum ./
COPY benchmarking-utilities/go/utils/ ./utils/
COPY docker/probes/go/ ./cmd/hhe-go-probe/
RUN mkdir -p /out \
 && CGO_ENABLED=0 go build -trimpath -ldflags='-s -w -buildid=' -o /out/hhe-go-probe ./cmd/hhe-go-probe

FROM ${GO_IMAGE} AS orchestrator-builder
WORKDIR /src
COPY benchmarking-utilities/orchestrator/go.mod ./
COPY benchmarking-utilities/orchestrator/main.go ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w -buildid=' -o /out-hhe .

FROM ${RUST_IMAGE} AS rust-builder
WORKDIR /src
COPY benchmarking-utilities/rust/Cargo.toml benchmarking-utilities/rust/Cargo.lock ./
COPY benchmarking-utilities/rust/lib.rs ./lib.rs
COPY benchmarking-utilities/rust/utils/ ./utils/
COPY docker/probes/rust/hhe-rust-probe.rs ./hhe-rust-probe.rs
RUN printf '\n[[bin]]\nname = "hhe-rust-probe"\npath = "hhe-rust-probe.rs"\n' >> Cargo.toml \
 && cargo build --locked --release --bin hhe-rust-probe \
 && cp target/release/hhe-rust-probe /out-hhe-rust-probe

FROM ${PYTHON_IMAGE} AS python-builder
ENV PIP_DISABLE_PIP_VERSION_CHECK=1 PIP_NO_CACHE_DIR=1
WORKDIR /build
COPY docker/python-requirements.lock ./requirements.lock
RUN python -m venv /opt/hhe-venv \
 && /opt/hhe-venv/bin/pip install --no-compile -r requirements.lock

FROM ${PYTHON_IMAGE} AS runtime
LABEL org.opencontainers.image.title="HHE-ARENA benchmark tooling" \
      org.opencontainers.image.description="Prebuilt cross-language benchmark utilities and report generator"
ENV PATH=/opt/hhe-venv/bin:/opt/hhe/bin:$PATH \
    PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1
WORKDIR /opt/hhe
COPY --from=python-builder /opt/hhe-venv /opt/hhe-venv
COPY --from=c-builder /out/hhe-c-schema-smoke /opt/hhe/bin/hhe-c-schema-smoke
COPY --from=c-builder /out/libhhe_benchmark_utils.a /opt/hhe/lib/libhhe_benchmark_utils.a
COPY --from=go-builder /out/hhe-go-probe /opt/hhe/bin/hhe-go-probe
COPY --from=orchestrator-builder /out-hhe /opt/hhe/bin/hhe
COPY --from=rust-builder /out-hhe-rust-probe /opt/hhe/bin/hhe-rust-probe
COPY benchmarking-utilities/report/report_generator.py /opt/hhe/bin/hhe-report
COPY benchmarking-utilities/python/utils/ /opt/hhe/python-utils/
COPY schemas/benchmark.schema.json /opt/hhe/schemas/benchmark.schema.json
COPY docker/entrypoint.sh /opt/hhe/bin/hhe-tooling
COPY docker/preflight.sh /opt/hhe/bin/hhe-preflight
COPY docker/benchmark-manifest.json /opt/hhe/benchmarks/manifest.json
RUN chmod 0555 /opt/hhe/bin/* \
 && mkdir -p /work \
 && python -c 'import cffi, jsonschema, psutil' \
 && hhe help >/dev/null \
 && hhe-report --help >/dev/null \
 && hhe-go-probe >/dev/null \
 && hhe-rust-probe >/dev/null
WORKDIR /work
ENTRYPOINT ["hhe-tooling"]
CMD ["--help"]
