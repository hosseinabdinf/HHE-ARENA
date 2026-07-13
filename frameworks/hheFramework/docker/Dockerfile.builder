# syntax=docker/dockerfile:1.7
FROM gcc:14.2-bookworm@sha256:b99b86a28812b1e6453a231a947dc43d76fe192788a12f344a9b568bf9f5d24c

ARG DEBIAN_FRONTEND=noninteractive
RUN apt-get update \
 && apt-get install -y --no-install-recommends \
      autoconf \
      automake \
      ca-certificates \
      cmake \
      git \
      libgmp-dev \
      libtool \
      m4 \
      make \
      pkg-config \
      wget \
      xz-utils \
 && rm -rf /var/lib/apt/lists/*

WORKDIR /workspace
