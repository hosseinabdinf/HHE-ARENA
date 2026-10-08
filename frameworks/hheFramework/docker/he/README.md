# HHE Framework HE/HHE package

The image builds HElib at `f0e3e010009c592cd411ba96baa8376eb485247a`,
SEAL at `265d77e84fb316cc9ed88cda895e8964e802e854`, and TFHE at
`9373b5d8a5b022ca5b4e112ae1c2440bc18e160c`. M4RI, NTL, GMP, and
other direct Debian build dependencies are version pinned in the Dockerfile.
The runtime inherits the compiler-free tooling image and contains 55 backend
executables for 217 manifest jobs. The source project uses C++17 and the
configured TFHE library includes FMA support; use an x86-64 CPU with the
required instruction set for functional runs. Source: the checked-in
`frameworks/hheFramework` implementation and its HElib, SEAL, and TFHE
upstreams. Consult their licenses before redistribution.

Build and perform workload-free checks from the repository root:

```sh
docker build --pull=false -t hhe-arena:tooling .
docker build --pull=false -f frameworks/hheFramework/docker/he/Dockerfile \
  -t hhe-arena:hhe-framework-he-matrix2 .
docker run --rm --network none hhe-arena:hhe-framework-he-matrix2 \
  hhe-validate-framework-image --require-benchmarks --check-links
docker run --rm --network none hhe-arena:hhe-framework-he-matrix2 list
```

These commands do not start an HE/HHE workload. The matrix keeps every HE/HHE
target deferred until the exact variant, backend, and operation has bounded
functional evidence on suitable hardware. HE/HHE evaluation can consume many
GiB and take substantial time. Set an explicit memory and time limit for each
candidate before a direct container `bench` run. Record raw JSONL and the
report separately from comparable performance results. Use
`docs/release-protocol.md` for strict host and security-profile requirements.
