# Fast Transciphering Via Batched And Reconfigurable LUT Evaluation

This repository contains the materials necessary to replicate the timings of the article:

“Fast Transciphering Via Batched And Reconfigurable LUT Evaluation”. IACR Transactions on Cryptographic Hardware and Embedded Systems, vol. 2024, no. 4, Sept. 2024, pp. 205-30, https://doi.org/10.46586/tches.v2024.i4.205-230.

# Build

Build instructions:

- Install Clang Version >= 15 and set the ```CC``` and ```CXX``` variables accordingly.
- chmod +x ```install_openfhe.sh```
- ```./install_openfhe.sh 64```. Replace 64 by 32 for Paramter Set I or II, and change the set as well in LUTEvalParams.h when using a set different from Set IV
- mkdir build && cd build
- cmake ..
- make -j 16

# Execution

For optimal performance during execution, we suggest the following:

- Important, especially on laptops. Determine which CPU cores are performance cores (as opposed to E cores), e.g. via a system monitor
- Use ```cpufreq-set``` to the the cpu frequency to maximum. By default, processors will not run at the maximum frequency.
- Several ```example_*``` should be in the build directory. The most important binaries are ```example_aes``` and ```example_bitdecomp``` (for LUT timing)
- Execute using ```taskset``` and assign the core which was set to maximum frequency

# Related works

The folder related_work_scripts contains two scripts. The first one contains a sage script using the lattice estimator
to estimate the security of the first related work on transciphering (TSBS23). The second script estimates the failure probability of related work number two (WWL24).

# Docker smoke run

From this directory, build and run the `example_aes` executable with:

```sh
docker build -f docker/Dockerfile -t hhe-arena-fbt .
docker run --rm --memory 8g --cpus 2 hhe-arena-fbt
```

The image rebuilds the pinned OpenFHE 1.1.1 commit with FbT's checked-in
patches, then runs one key-scheduler measurement with no warmup before the
example's single AES transcipher call. The generated JSONL stays in `/tmp`
inside the disposable container.

The image build has been verified. The default SET_IV execution exceeded the
8 GiB guard and was OOM-killed with exit code 137 before transciphering
completed. Increase the container memory limit only when the host has enough
capacity; the limit in the command above is intentionally conservative.
