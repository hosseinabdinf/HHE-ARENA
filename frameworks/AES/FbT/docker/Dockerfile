FROM ubuntu:24.04

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        ca-certificates \
        clang \
        cmake \
        git \
        make \
        patch \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /opt/fbt
COPY patches/ patches/

# Recreate the patched OpenFHE install because the checked-in package contains
# only part of the required header tree.
RUN rm -rf libs/openfhe /tmp/openfhe /opt/openfhe-build /opt/fbt-build \
    && git clone https://github.com/openfheorg/openfhe-development.git \
        /tmp/openfhe \
    && git -C /tmp/openfhe checkout --detach 4ebb28e \
    && cd /tmp/openfhe/src/binfhe \
    && patch -s -p0 < /opt/fbt/patches/include.patch \
    && patch -s -p0 < /opt/fbt/patches/lib_noiseless.patch \
    && rm -rf /opt/openfhe-build /tmp/openfhe/CMakeCache.txt \
        /tmp/openfhe/CMakeFiles \
    && cmake -S /tmp/openfhe -B /opt/openfhe-build \
        -DCMAKE_BUILD_TYPE=Release \
        -DCMAKE_C_COMPILER=clang \
        -DCMAKE_CXX_COMPILER=clang++ \
        -DCMAKE_INSTALL_PREFIX=/opt/fbt/libs/openfhe \
        -DNATIVE_SIZE=64 \
        -DWITH_NATIVEOPT=ON \
        -DBUILD_STATIC=ON \
        -DBUILD_SHARED=OFF \
        -DWITH_OPENMP=OFF \
    && cmake --build /opt/openfhe-build --parallel 2 \
    && cmake --install /opt/openfhe-build \
    && cd /opt/fbt \
    && rm -rf /tmp/openfhe /opt/openfhe-build
COPY . .


RUN cmake -S /opt/fbt -B /opt/fbt-build \
        -DCMAKE_BUILD_TYPE=Release \
        -DCMAKE_C_COMPILER=clang \
        -DCMAKE_CXX_COMPILER=clang++ \
        -DBUILD_STATIC=ON \
        -DFBT_ENABLE_PROFILE_GENERATION=OFF \
    && cmake --build /opt/fbt-build --target example_aes --parallel 2

ENV HHE_BENCH_CIPHER=fbt-aes \
    HHE_BENCH_BACKEND=openfhe \
    HHE_BENCH_IMPLEMENTATION=fbt-example-aes \
    HHE_BENCH_ITERATIONS=1 \
    HHE_BENCH_WARMUPS=0 \
    HHE_BENCH_POLICY=warn \
    HHE_BENCH_OUTPUT=/tmp/fbt-example-aes.jsonl

ENTRYPOINT ["/opt/fbt-build/example_aes"]
