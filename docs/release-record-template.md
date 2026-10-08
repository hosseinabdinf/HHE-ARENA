# HHE-ARENA release evidence record

Copy this file into the release archive and replace every `TODO`. Keep the
archived JSONL, reports, manifests, host metadata, and checksum file beside it.
Do not mark a target supported from JSONL schema validity alone.

## Release identity

| Field | Value |
|---|---|
| Release name/version | TODO |
| Source commit | TODO |
| Working tree clean | TODO |
| Matrix version and status | `2 / draft` until formally frozen |
| Approved scope decision | TODO |
| Release date | TODO |
| Archive checksum file | TODO |

## Scope and exclusions

| Item | Decision/evidence |
|---|---|
| Included catalog image IDs | TODO |
| Included ready matrix targets | TODO |
| Deferred targets excluded from support claims | TODO |
| Dependency-blocked targets and blockers | TODO |
| Other exclusions (including AES/FbT if unchanged) | TODO |
| Security profile claims and supporting evidence | TODO |

## Per-target evidence

Add one row per included `variant/backend/operation` target. Link paths relative
to this record. Use the registry digest, not a local Docker image ID.

| Variant ID | Image ID | Registry digest | Manifest job ID | Backend / operation | Correctness result | Raw JSONL | Report | SHA-256 list | Limitations |
|---|---|---|---|---|---|---|---|---|---|
| TODO | TODO | TODO | TODO | TODO | TODO | TODO | TODO | TODO | TODO |

## Functional run metadata

| Field | Value |
|---|---|
| Host ID and platform | TODO |
| Preflight command and result | TODO |
| CPU model/topology and selected CPU | TODO |
| Kernel and OS | TODO |
| RDTSCP/invariant TSC | TODO |
| Governor and Turbo state | TODO |
| RAM and swap | TODO |
| Docker version | TODO |
| Image digest and source revision | TODO |
| Iteration/warmup policy | TODO |

## Comparable performance evidence

Complete only when strict preflight passes on an approved bare-metal host.

| Field | Value |
|---|---|
| Protocol version | TODO |
| Security profile | TODO |
| CPU isolation and sibling-thread state | TODO |
| Invocation order and repetition IDs | TODO |
| Raw per-run records and pooled report | TODO |
| Deviations from `docs/release-protocol.md` | TODO |

## Release checks

- [ ] Matrix/catalog/manifest schema and relationship validators pass.
- [ ] Every included target has exact selector and correctness evidence.
- [ ] Every image is published and pullable by immutable digest.
- [ ] Deferred and blocked targets are listed with reasons.
- [ ] Security labels have supporting audit evidence.
- [ ] Strict host preflight passed for performance claims.
- [ ] Archive files and this record have SHA-256 checksums.
- [ ] Matrix v2 scope and status are approved for freezing.

## Review

| Role | Name | Decision | Date |
|---|---|---|---|
| Scope owner | TODO | TODO | TODO |
| Security review | TODO | TODO | TODO |
| Release review | TODO | TODO | TODO |
