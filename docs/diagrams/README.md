# Diagrams

These are the two deployment shapes this repo can start for you. Same Eshu,
different wrapping. If you want the ideas in plain words first, read
[Concepts](../concepts/README.md). If you want to run one, start at
[Getting started](../getting-started/README.md).

## On your laptop with Compose

*Figure 1: the default Compose stack in `compose/docker-compose.yaml`.*

```mermaid
flowchart LR
  Repos["Your repositories"]
  Setup["One-shot setup: db-migrate, workspace-setup, bootstrap-index"]
  Ingester["ingester"]
  Pg[("postgres: facts and queues")]
  Reducer["resolution-engine (reducer)"]
  Projector["projector"]
  Graph[("nornicdb: graph")]
  API["eshu API :8080"]
  MCP["mcp-server :8081"]
  You["You and your assistants"]
  Setup --> Pg
  Setup --> Graph
  Repos --> Ingester
  Ingester -->|"facts"| Pg
  Pg -->|"queued work"| Reducer
  Pg --> Projector
  Reducer -->|"linked graph"| Graph
  You --> API
  You --> MCP
  API --> Pg
  API --> Graph
  MCP --> Pg
  MCP --> Graph
```

In words: the one-shot jobs prepare Postgres and the graph, then the ingester
keeps feeding facts into Postgres while the reducer links them into NornicDB.
You query the API on port 8080 or the MCP server on port 8081. The projector
runs alongside and drains source-local projection work the first pass left
behind.

## As a shared service with Helm

*Figure 2: the Helm chart in `helm/eshu`, rendering separate workloads.*

```mermaid
flowchart LR
  You["You: values file and Secrets"]
  Helm["Helm or Argo CD"]
  Job["Schema bootstrap Job (runs first)"]
  API["API Deployment"]
  MCP["MCP Deployment"]
  Ingester["Ingester StatefulSet + PVC"]
  Reducer["Resolution Engine Deployment"]
  Extra["Optional: workflow coordinator, collectors, webhook"]
  Pg[("Your Postgres (chart does not install it)")]
  Graph[("Bolt graph backend (yours, or chart NornicDB)")]
  You --> Helm
  Helm --> Job
  Helm --> API
  Helm --> MCP
  Helm --> Ingester
  Helm --> Reducer
  Helm -.-> Extra
  Job --> Pg
  Job --> Graph
  Ingester -->|"facts"| Pg
  Pg -->|"queued work"| Reducer
  Reducer -->|"linked graph"| Graph
  Extra -->|"facts"| Pg
  API --> Pg
  API --> Graph
  MCP --> Pg
  MCP --> Graph
```

In words: Helm runs the schema bootstrap Job first, then rolls out the API,
MCP server, ingester, and resolution engine as separate workloads. Postgres
is yours to provide, because the chart doesn't install it. The graph backend
is usually an endpoint you point at, though the chart can run NornicDB for you
if you ask it to.

## Related diagrams

- The EKS target shape is drawn in [Deploy To EKS](../eks/index.md).
- The "pick your path" choice is drawn in the [repo README](../../README.md).
- Why the shapes look like this: [Architecture](../architecture/README.md).
