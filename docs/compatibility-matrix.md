# Compatibility Matrix

This matrix supersedes `docs/kubernetes/runtime-parity-matrix.md` as the
published cross-repository contract. It records the tested deployment
artifact pairs. Independent releases widen the range only with tested pairs
recorded here.

Run the verifier from the repository root:

```bash
scripts/verify-compose-helm-runtime-parity.sh
```

## Tested pairs (v0)

Single tested pair at extraction, taken from the `eshu-hq/eshu#8871`
deployment contract (chart `eshu` `0.0.3-pre-release-26`, appVersion
`v0.0.3-pre-release-18`). Oldest and newest converge on that pair until a
later release records a wider tested range.

| Component | Oldest supported | Newest tested |
| --- | --- | --- |
| eshu core (`ghcr.io/eshu-hq/eshu`) | `v0.0.3-pre-release-18` | `v0.0.3-pre-release-18` |
| eshu chart (`eshu`, `version: 0.0.3-pre-release-26`) | `0.0.3-pre-release-26` | `0.0.3-pre-release-26` |
| Loki collector (digest-pinned) | `sha256:12eac89e7abd2988529c6f2854ba63293b5507502630764c3f6b8ffc7e3bf84d` | `sha256:12eac89e7abd2988529c6f2854ba63293b5507502630764c3f6b8ffc7e3bf84d` |
| Tempo collector (digest-pinned) | `sha256:0230e174284bb7a9831babac9b9927be05df02e8620c20e448e91dbb1ac2ecac` | `sha256:0230e174284bb7a9831babac9b9927be05df02e8620c20e448e91dbb1ac2ecac` |
| Postgres | `postgres:18-alpine` | `postgres:18-alpine` |
| Neo4j (compose default) | `neo4j:2026-community@sha256:eabfbb042bdaca2fd5e1950db1329b22c794eee80f0eacc4e7a729d44b2e863f` | `neo4j:2026-community@sha256:eabfbb042bdaca2fd5e1950db1329b22c794eee80f0eacc4e7a729d44b2e863f` |
| NornicDB (`nornicdb-amd64-cpu`) | `fix-500-e022384c@sha256:74a8ed7b36f37bdd1a7e32d8bc6aa3fa88908b7207bfa6568567ab94e4a4b3b1` | `fix-500-e022384c@sha256:74a8ed7b36f37bdd1a7e32d8bc6aa3fa88908b7207bfa6568567ab94e4a4b3b1` |

Credential handling: ArgoCD overlays carry secret-reference examples only,
no credential material; claim-driven collectors mount redaction keys at
runtime. Artifact provenance: release image from the core `Dockerfile` in
`eshu-hq/eshu`; chart published via the `chart-release` lane to
`oci://ghcr.io/<owner>/charts`. Vulnerability response ownership is named
in `SECURITY.md`.
