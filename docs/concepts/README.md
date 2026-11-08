# Concepts

The deployment ideas in plain words. No commands here. When you're ready to
run something, go to [Getting started](../getting-started/README.md). When
you want the pictures, see [Diagrams](../diagrams/README.md).

## This repo starts Eshu. It isn't Eshu

This repo doesn't contain Eshu itself. It contains the files that start Eshu.
You bring the settings, the secrets, and somewhere to run it. The Compose
files or the Helm chart here turn that into running services. If you want to
change how Eshu behaves, you're in the wrong repo. That code lives in
[eshu](https://github.com/eshu-hq/eshu).

## The shape stays the same

Whichever path you pick, the shape stays the same. A short-lived setup job
prepares the databases first. Then the ingester reads your repositories and
stores facts in Postgres. Facts are small, checked records of what was seen.
The reducer, also called the resolution engine, links those facts into the
graph. You and your assistants read the result through the API and the MCP
server. Collectors are extra. They stay off until you enable them, and each
one watches a single outside system.

## Two topologies, one idea

- **On your laptop (Compose).** One Compose project runs everything
  together: Postgres, the graph, the setup jobs, the ingester, the reducer,
  a projector, the API on port 8080, and the MCP server on port 8081. The
  services build from the `Dockerfile` in your `eshu` checkout, so you need
  that checkout nearby. Other files in `compose/` cover a Neo4j
  compatibility stack, local telemetry, a demo, and test stacks.
- **As a shared service (Helm).** The chart renders separate workloads
  instead of one project: a schema bootstrap Job that runs first, then API,
  MCP, ingester, and resolution engine Deployments, with the workflow
  coordinator, collectors, and webhook as options. EKS and Argo CD are this
  same Helm shape with different wrapping.

Both shapes are drawn, with text versions, in
[Diagrams](../diagrams/README.md).

## Words worth keeping straight

- **Healthy isn't finished.** Health tells you a process is up. It doesn't
  tell you indexing finished. Empty answers usually mean the setup job, the
  ingester, or the reducer still has work to do.
- **Collectors only watch.** A collector reads one outside system and emits
  facts. Leave it off and nothing here goes near that system.
- **Only the reducer writes linked truth.** Everything else hands facts to
  Postgres. The reducer is the one step that joins them into the graph.

## Next

- [Architecture](../architecture/README.md) — what this repo owns, and what it never does.
- [Getting started](../getting-started/README.md) — the shortest path to a running stack.
