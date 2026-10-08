# Available schemes and benchmark images

This page lists the scheme implementations packaged in
`manifests/images.v2.json`. Matrix-v2 execution readiness varies by exact
selector; catalog presence alone does not make a deferred operation runnable.

“Packaged” means that the implementation has a catalog entry, a version-1
benchmark manifest, a Docker build definition, a non-root runtime contract,
and declared executable jobs. It does not mean that every matrix selector is
ready, that a workload has already been run, or that measurements are
publication-grade. Check `docs/support-matrix.md` for selector readiness,
`results/` for collected runs, and `docs/status.md` for migration and dependency
notes.

Run a registered image through the catalog CLI, for example:

```text
hhe bench --catalog manifests/images.v2.json --image-policy auto \
  --framework AES/Hippogryph --cipher "AES-128 Hippogryph" \
  --operation "Transciphering" --iterations 1 --warmup 0
```

The catalog currently contains 26 image entries. Matrix readiness is tracked
per exact operation in the support matrix.

| Framework / path | Scheme or cipher | Docker image | Contract manifest | Declared operations |
|---|---|---|---|---|
| AES/OpenSSL | AES-128-CBC | `hhe-arena:openssl-aes-matrix1` | [openssl-aes.v1.json](../manifests/openssl-aes.v1.json) | AES-128-CBC Encryption |
| AES/TinyAES | AES-128-CBC | `hhe-arena:tinyaes-aes-matrix2` | [tinyaes-aes.v2.json](../manifests/tinyaes-aes.v2.json) | AES-128-CBC Encryption |
| AES/Full-LUT | AES-128 full-LUT | `hhe-arena:full-lut-aes-matrix2` | [full-lut-aes.v2.json](../manifests/full-lut-aes.v2.json) | Key expansion; TFHE symmetric encryption; transciphering |
| AES/Fregata CBS | AES-128 CBS | `hhe-arena:fregata-cbs-aes-matrix2` | [fregata-cbs-aes.v2.json](../manifests/fregata-cbs-aes.v2.json) | AES and TFHE++ boundaries |
| AES/Fregata GBS | AES-128 GBS | `hhe-arena:fregata-gbs-aes-matrix2` | [fregata-gbs-aes.v2.json](../manifests/fregata-gbs-aes.v2.json) | GBS AES pipeline |
| AES/FFT-CBS | AES, half-CBS, CBS, HP-LHE | `hhe-arena:fft-cbs-aes-matrix2` | [fft-cbs-aes.v2.json](../manifests/fft-cbs-aes.v2.json) | SKE and HHE/evaluation boundaries |
| AES/Hippogryph | AES-128 Hippogryph | `hhe-arena:hippogryph-aes-matrix2` | [hippogryph-aes.v2.json](../manifests/hippogryph-aes.v2.json) | TFHE plaintext encryption; transciphering |
| AES/P-encoding | AES-128 P-encoding | `hhe-arena:p-encoding-aes-matrix2` | [p-encoding-aes.v2.json](../manifests/p-encoding-aes.v2.json) | TFHE plaintext encryption; transciphering |
| HHELand | Rubato-128-L | `hhe-arena:hheland-rubato-matrix1` | [hheland-rubato.v1.json](../manifests/hheland-rubato.v1.json) | SKE.Enc() |
| HHELand | Hera and Pasta | `hhe-arena:hheland-symmetric-matrix2` | [hheland-symmetric.v2.json](../manifests/hheland-symmetric.v2.json) | SKE.Enc() |
| HHELand | HHE, RtF, BGV, and CKKS | `hhe-arena:hheland-hhe-matrix2` | [hheland-hhe-he.v2.json](../manifests/hheland-hhe-he.v2.json) | HHE.Decomp(); RtF; HE encrypt/evaluate |
| Fasta/Rasta | Fasta | `hhe-arena:fasta-symmetric-matrix2` | [fasta-symmetric.v2.json](../manifests/fasta-symmetric.v2.json) | SKE encrypt; SKE decrypt |
| Fasta/Rasta | Fasta and Rasta | `hhe-arena:fasta-rasta-matrix2` | [fasta-rasta.v2.json](../manifests/fasta-rasta.v2.json) | HElib/TFHE setup; HHE transciphering |
| HHE Framework | 23 plain cipher variants | `hhe-arena:hhe-framework-plain-matrix2` | [hhe-framework-plain.v2.json](../manifests/hhe-framework-plain.v2.json) | SKE encrypt; SKE decrypt; plain evaluation |
| Elisabeth | Elisabeth clear | `hhe-arena:elisabeth-clear-matrix1` | [elisabeth-clear.v1.json](../manifests/elisabeth-clear.v1.json) | SKE.Enc(); SKE.Dec() |
| Elisabeth | Elisabeth HHE | `hhe-arena:elisabeth-hhe-matrix2` | [elisabeth-hhe.v2.json](../manifests/elisabeth-hhe.v2.json) | HHE.Decomp |
| Transistor | Transistor symmetric | `hhe-arena:transistor-symmetric-matrix1` | [transistor-symmetric.v1.json](../manifests/transistor-symmetric.v1.json) | Sym.Enc() |
| Transistor | Transistor homomorphic | `hhe-arena:transistor-homomorphic-matrix2` | [transistor-homomorphic.v2.json](../manifests/transistor-homomorphic.v2.json) | HHE.Decomp() |
| Margrethe | Margrethe SKE | `hhe-arena:margrethe-symmetric-matrix2` | [margrethe-symmetric.v2.json](../manifests/margrethe-symmetric.v2.json) | Margrethe SKE |
| Margrethe | Margrethe HHE | `hhe-arena:margrethe-hhe-matrix2` | [margrethe-hhe.v2.json](../manifests/margrethe-hhe.v2.json) | HHE.Decomp() |
| FRAST | FRAST | `hhe-arena:frast-matrix2` | [frast.v2.json](../manifests/frast.v2.json) | SKE, setup, KS, PBS, online, and HHE boundaries |
| Trivium/Kreyvium | Trivium and Kreyvium | `hhe-arena:trkr-matrix2` | [trkr.v2.json](../manifests/trkr.v2.json) | Boolean, byte, short-integer, and transencrypt operations |
| FiLIP/independent-PTX | FiLIP | `hhe-arena:filip-independent-ptx-matrix2` | [filip-independent-ptx.v2.json](../manifests/filip-independent-ptx.v2.json) | HHE setup; SKE.Enc(); HHE.Decomp() |
| FiLIP/Sorting Hat | FiLIP | `hhe-arena:filip-sorting-hat-matrix2` | [filip-sorting-hat.v2.json](../manifests/filip-sorting-hat.v2.json) | SKE.Enc(); HHE.Decomp() |
| YuX | Yu2x and Yupx-p | `hhe-arena:yux-symmetric-matrix2` | [yux-symmetric.v2.json](../manifests/yux-symmetric.v2.json) | Key expansion; encryption; decryption |

For each entry, inspect the manifest with `hhe list`, use
`hhe bench --dry-run` to review dispatch, and begin with a small warn-policy
run. HE/HHE workloads may require substantial memory and suitable Linux
hardware; WSL timing results should be treated as functional evidence rather
than publication data.
