# Supported implementation matrix

The machine-readable support declaration is
[`manifests/support-matrix.v1.json`](../manifests/support-matrix.v1.json),
validated by [`schemas/support-matrix.schema.json`](../schemas/support-matrix.schema.json).
Matrix version 1 is frozen as the first Docker and production-manifest slice.
All four selected entries now have verified contract-v1 images and manifests;
frameworks outside this slice remain later expansion candidates.

The expansion inventory is
[`manifests/support-matrix.v2.json`](../manifests/support-matrix.v2.json).
Version 2 is deliberately **draft**: it records all migrated, project-owned
non-FbT candidate paths and their packaging blockers, while
`images.v2.json` contains only the subset with validated runnable packaging.
The frozen v1 files remain the authoritative baseline slice.

## Version 1 selection

| Priority | Framework | Language | Selected safe path | Backend | Resource | Dispatch readiness |
|---:|---|---|---|---|---|---|
| 1 | AES/OpenSSL | C | AES-128-CBC encryption | OpenSSL | light | Image and one-job manifest ready. |
| 2 | HHELand | Go | Rubato-128-L `SKE.Enc()` | native | light | Go 1.24 image and one-job manifest ready. |
| 3 | Elisabeth | Rust | clear `SKE.Enc()` / `SKE.Dec()` | native | light | Stable operation-selectable binary and two-job manifest ready. |
| 4 | Transistor | Python | `Sym.Enc()` | native | light | Fixed launcher, precompiled counter, Sage-free GF(17) fallback, and known-answer image check ready. |

These four entries exercise all maintained language harnesses without running
HE or HHE work. AES/FbT remains explicitly excluded. The security profile is
left null unless the selected parameters have been verified against a named
profile; language coverage must not be confused with security normalization.

## Meaning of statuses

- `selection_status=selected` means the implementation belongs to the frozen
  version-1 delivery slice.
- `docker_status=ready` requires a pinned multi-stage image that contains the
  fixed `container_executable` and no build toolchain in its runtime layer.
- `manifest_status=ready` requires one concrete manifest-v1 job per operation,
  with the executable path checked during the image build.
- `safe_smoke=true` authorizes only a lightweight symmetric smoke after the
  normal project checks. It does not authorize any HE/HHE test or benchmark.

Changes to membership or executable contracts require a new matrix version.
Progress-only fields may be updated in version 1 as Docker and manifest work
advances.

## Draft version 2 inventory and selector taxonomy

Version 2 separates two concepts that must not be conflated:

- 29 **implementation units** describe Docker/build/manifest readiness.
- 103 **benchmark variants** describe the cipher and parameter choices that
  the final CLI must expose.

The normalized selector order is
`family -> path -> cipher -> variant -> framework -> backend -> language`.
Every level accepts `all`. The families are `aes` (the repository's
AES/conventional collection, including its SM4 implementations),
`he_friendly`, and `pure_he`; paths are `ske`, `hhe`, and `he`.
Pure-HE primitive profiles use `he`, not `hhe`. A variant may expose more
than one path when the same implementation has both a native symmetric
benchmark and an HE-backed transciphering benchmark.

This permits selections such as:

```text
family=aes         path=ske cipher=AES     variant=all
family=he_friendly path=ske cipher=Rubato  variant=128S
family=he_friendly path=hhe cipher=Rubato  variant=all
family=all         path=all cipher=all     variant=all
```

The draft contains 10 AES/conventional variants, 72 HE-friendly variants, and
21 pure-HE variants. By path it exposes 55 SKE targets, 76 HHE targets, and 21
HE targets; variants with both SKE and HHE paths count in both target totals.
Notable parameterized groups include all six HHELand Rubato profiles, eight
Hera profiles, six Pasta profiles, 25 HHE Framework cipher configurations,
12 FFT-CBS AES/CBS/HP-LHE profiles, six trkr representations, and all eight
YuX benchmark executables. Commented alternatives and dependency-owned
primitive suites are intentionally not advertised as runnable legacy
benchmarks.

The 29 implementation units have these readiness classes after splitting
native Fasta, HHE Framework plain/SKE, and YuX symmetric from their HE/HHE
packages:

| Readiness class | Count | Meaning |
|---|---:|---|
| `safe_lightweight` | 10 | A non-HE path can be bounded-smoked and packaged independently. |
| `mixed` | 0 | Every safe clear path formerly grouped with HE/HHE work now has a separate packaging unit. |
| `he_hhe_heavy` | 18 | The migrated boundary exists, but execution is not authorized as a routine smoke and dependency/resource packaging remains. |
| `dependency_blocked` | 1 | The FiLIP Rust reference cannot currently build because its required upstream branch is unavailable. |

The inventory excludes AES/FbT, dependency-owned primitive suites, copied
Lattigo/ring microbenchmarks, vendored TFHE tests, and Fregata's primitive
Google Benchmark suite. Those exclusions are scope decisions, not forgotten
image entries.

The host CLI now consumes this hierarchy in `list`, `bench`, and
`bench-all`. With `images.v2.json`, it automatically joins the sibling
support matrix to the catalog by `implementation_id` and operation. Listing
shows both the declared target status and whether it is effectively runnable;
execution is limited to targets declared `ready` with a matching real image
entry. Deferred, blocked, and uncataloged variants remain visible but cannot
execute. A target may override its variant-level `implementation_id` when
lightweight and HE/HHE paths are intentionally packaged as separate images.

[`manifests/images.v2.json`](../manifests/images.v2.json) now contains thirteen
validated images: the four retained v1 units, Transistor homomorphic Rust,
TinyAES, HHELand symmetric, Margrethe SKE, Fasta native, HHE Framework plain,
YuX symmetric, Full-LUT AES/TFHE, and Fregata CBS. Fregata adds two native AES
and three TFHE++ jobs, bringing the current CLI to 112 runnable operation
choices. Heavy operations remain opt-in and are not routine smokes.

An image-catalog entry is added only after its candidate has a pinned
multi-stage Docker recipe, a compiler-free non-root runtime, a concrete
embedded manifest, and contract validation. The next packaging order is:

1. continue build/contract validation with Fregata GBS AES and later heavy
   candidates without running them;
2. HE/HHE functional validation only with explicit authorization.
