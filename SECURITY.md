# Security Policy

This repository publishes Eshu deployment assets (Helm chart, Compose lanes,
GitOps examples, deployment defaults, compatibility matrices). The release
image, the schema-bootstrap and repair implementation, and every Go binary
stay core-owned by `eshu-hq/eshu`.

## Supported Versions

Security fixes target the latest release only.

| Version | Supported |
| ------- | --------- |
| Latest  | Yes       |
| Older   | No        |

## Vulnerability response ownership

Accountable owner for this stream: [@linuxdynasty](https://github.com/linuxdynasty)
(owner-confirmed in `eshu-hq/eshu#8871`). Chart and image-pin updates that fix
a deployment-surface vulnerability land here through the chart release lane
(`.github/workflows/chart-release.yml`); fixes to the release image or to the
schema-bootstrap and repair implementation land in `eshu-hq/eshu` and are
picked up here as a new pinned release.

## Reporting a Vulnerability

Do not open a public GitHub issue for security problems.

Instead:

1. Use [GitHub private vulnerability reporting](https://docs.github.com/en/code-security/security-advisories/guidance-on-reporting-and-writing-information-about-vulnerabilities/privately-reporting-a-security-vulnerability) if enabled for this repository.
2. If private reporting is unavailable, contact the maintainers privately before disclosure.
3. Include the impact, reproduction steps, affected components, and any suggested mitigation.

We aim to acknowledge reports promptly and coordinate disclosure after a fix is available.

## Disclosure Guidelines

- Do not publicly disclose an issue before a fix is released.
- Avoid testing in ways that could harm users, systems, or data.
- Report in good faith with enough detail to reproduce the issue.

## Operational Notes

When running Eshu in production:

- Keep dependencies up to date.
- Restrict network exposure of HTTP endpoints — the API is designed for internal use.
- Commit no credentials: ArgoCD overlays and minimal manifests carry
  secret-reference examples only, never credential material.
- Review graph data sources and access boundaries, especially when indexing repos from multiple orgs.
