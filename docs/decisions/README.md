# Decisions

Architecture decision records (ADRs) for this deployment repo will live
here. There are none yet, and this page won't pretend otherwise. It sets
the format so the first real record has a home.

## Index

No decision records yet.

## Format for a new record

When a deployment choice is expensive to reverse — a chart contract, a
storage promise, a supported upgrade path — write it down here. Keep it
short and plain:

1. **Context** — what forced the choice, in a few sentences.
2. **Decision** — what we chose, in one clear statement.
3. **Consequences** — what gets easier, what gets harder, and what we
   accepted knowingly.
4. **Alternatives** — what we considered and why we did not pick it.
5. **Status** — proposed, accepted, or superseded, with a link to whatever
   replaced it.

Name the file `NNNN-short-title.md` with the next free number, add it to
the index above, and link it from the pull request that acts on it.

## Nearby records that are not ADRs

- [Seed Provenance](../seed-provenance.md) records where these files came
  from. It's a provenance note, not a decision record.
- [Compatibility Matrix](../compatibility-matrix.md) records tested version
  pairs. It's a contract, updated as pairs are proven.
- Eshu's own design decisions live with the core in
  [eshu](https://github.com/eshu-hq/eshu/blob/main/docs/public/decisions/index.md).
  Deployment decisions that belong to the programs belong there, not here.
