# Framework for Hybrid Homomorphic Encryption

This repository contains code to test and benchmark symmetric ciphers in hybrid homomorphic encryption. This code is also used in the evaluations of the ciphers in the paper [1].

## HE libraries

The following Homomorphic Encryption Libraries are part of the framework:

- [HElib](https://github.com/homenc/HElib/) (version 2.2.2)
- [SEAL](https://github.com/Microsoft/SEAL/) (version 3.7.3)
- [TFHE](https://github.com/tfhe/tfhe) (version 1.1)

The libraries are included as submodules in this repository and will be downloaded to the `thirdparty` directory.

## Ciphers

The following ciphers are already implemented in the framework:

- [LowMC](https://eprint.iacr.org/2016/687.pdf)
- [Rasta](https://eprint.iacr.org/2018/181.pdf)
- [Agrasta](https://eprint.iacr.org/2018/181.pdf)
- [Dasta](https://tosc.iacr.org/index.php/ToSC/article/view/8696/8288)
- [Kreyvium](https://eprint.iacr.org/2015/113.pdf)
- [FiLIP](https://eprint.iacr.org/2019/483.pdf)
- [Masta](https://ieeexplore.ieee.org/stamp/stamp.jsp?tp=&arnumber=9240936)
- [Hera](https://eprint.iacr.org/2020/1335.pdf)
- [Pasta](https://eprint.iacr.org/2021/731.pdf)

## Compilation

The framework is developed and tested on Linux (and WSL2). The dependencies of the framework are `git`, `gcc/g++` (or `clang/clang++`), `cmake`, `autoconf`, and `libtool`. Install these dependencies with

```bash
sudo apt install build-essential cmake autoconf libtool
```

To compile the framework, execute the following commands from the root directory:

```bash
mkdir build
cd build
cmake ..
make -j4
```

Each cipher for each library will then have its own executable, located in `build/test`.

## Adding a cipher

A new cipher can be added to the framework by adding the source code to the `ciphers` directory. Make sure to use the same structure and naming convention as the already implemented ciphers! Furthermore, add the cipher to `CMakeLists.txt`.

## Benchmarking a cipher

The benchmarks are implemented in `ciphers/common/<lib>_kats` for binary
ciphers and `ciphers/common_Zp/<lib>_kats` for Z_p ciphers. Select a benchmark
by adding the corresponding known-answer test to the cipher's `testvectors.h`.

The generated test executables use the repository's shared schema-1.0.0 C/C++
harness. Each KAT is repeated over configured warm-up and measured phases,
while only named operations such as `ske_encrypt`, `he_key_encrypt`,
`hhe_decrypt`, `he_evaluate`, and `final_decrypt` are timed. Setup, console
output, and correctness checks remain outside each operation boundary.

For example, after building from the repository root:

```bash
mkdir -p results/hhe-framework
HHE_BENCH_CIPHER=pasta_3 \
HHE_BENCH_BACKEND=seal \
HHE_BENCH_IMPLEMENTATION=hhe-framework-pasta3-seal \
HHE_BENCH_ITERATIONS=3 \
HHE_BENCH_WARMUPS=1 \
HHE_BENCH_POLICY=warn \
HHE_BENCH_OUTPUT=results/hhe-framework/pasta3-seal.jsonl \
./build/tests/pasta_3_seal_test
```

`HHE_BENCH_CPU` optionally pins the process to a logical CPU.
`HHE_BENCH_SECURITY_PROFILE` records the normalized security profile. Use
`strict` policy only on a host prepared for publication-grade measurement.
Each executable may emit several operation groups to the same JSONL file.

## Docker benchmark CLI

From Windows PowerShell, run the Docker automation from the repository root:

```powershell
./frameworks/hheFramework/benchmark-docker.ps1
```

Choose a backend from the menu. The script discovers that backend's cipher list
from CMake, builds every corresponding test executable in the pinned Linux
builder image, runs each executable with the shared benchmark harness, and uses
the root tooling image to validate and report the combined JSONL dataset.

For a non-interactive plain run:

```powershell
./frameworks/hheFramework/benchmark-docker.ps1 -Backend plain -Iterations 10 -Warmups 2
```

Plain results are written to `results/hhe-framework/plain/raw.jsonl`, with
generated reports under `results/hhe-framework/plain/report/`. The first
framework configuration downloads and builds its third-party dependencies and
can take substantially longer than later runs. Use `-Policy warn` (the
default) for Docker Desktop functional measurements; virtualized cycle results
are not publication-grade.

## Citing our work

Please use the following BibTeX entry to cite our work in academic papers.

```tex
@article{HybridHE,
  author    = {Christoph Dobraunig and
               Lorenzo Grassi and
               Lukas Helminger and
               Christian Rechberger and
               Markus Schofnegger and
               Roman Walch},
  title     = {Pasta: A Case for Hybrid Homomorphic Encryption},
  journal   = {{IACR} Cryptol. ePrint Arch.},
  volume    = {2021},
  pages     = {731},
  year      = {2021}
}
```

# Acknowledgement

This work is partly supported by the European Union under the project Confidential6G with Grant agreement ID: 101096435.

[1] [https://eprint.iacr.org/2021/731.pdf](https://eprint.iacr.org/2021/731.pdf)
