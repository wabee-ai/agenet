# Agenet v2 — Migration from v1 and open questions

## Field-by-field mapping

| v1 | v2 |
|---|---|
| `uuid` | `messageId` on the first message (idempotency) and `taskId` assigned by the Gateway. A v1 shim uses the v1 uuid as both. |
| `requestType: "retrieve_file"` | `metadata["agenet.io/skill"] = "retrieve_file"`, or call the virtual endpoint `/agents/file-agent` |
| `content-type` + `data` | one `Part` with `mediaType` set; JSON becomes a `data` part, text a `text` part, binary a `raw` part |
| `metadata.clientId` | dropped; the principal comes from the token. The shim maps it to a client-credentials client id. |
| `metadata.timestamp` | `Task.status.timestamp`, set by the mesh |
| "message delivered" notification | `TaskStatusUpdateEvent{WORKING}` after agent `Ack` |
| partial response | `TaskArtifactUpdateEvent{append, lastChunk}` |
| "completed" | `TaskStatusUpdateEvent{COMPLETED}` |
| "error" | `TaskStatusUpdateEvent{FAILED, message}` with `google.rpc.Status` in `metadata["agenet.io/error"]` |
| Agent `connect(requestType)` | `Attach(Register{card})` with a skill per former requestType |
| Gateway client bidi stream | any A2A binding; streaming via `SendStreamingMessage` or `SubscribeToTask` |

## Migration path

1. **Shim first.** Ship a `v1-gateway` adapter that accepts the v1 JSON over the v1 gRPC
   stream and translates it to A2A against the v2 Gateway, and a `v1-agent` adapter that lets
   a v1 agent process register with a synthesized card (`skills=[{id: requestType}]`,
   `sideEffects: external`, unsigned card accepted only in a `legacy` trust tier). Both are
   Deprecated from day one with a 12-month removal date.
2. **Move agents.** Replace each v1 agent's loop with the v2 SDK `Attach` client; add a real
   card with schemas, side-effect hints, and a signature. Nothing on the client side changes.
3. **Move clients.** Point clients at a virtual endpoint using any A2A SDK, or at the mesh
   endpoint with the routing extension for dynamic skill selection.
4. **Turn on policy and budgets.** Start in `log_only` mode, review the audit log, then
   enforce.
5. **Remove the shims** after the deprecation window.

## Minimal reference implementation plan

Suggested order, each step independently useful:

1. Proto package `agenet.v2` importing `a2a.v1`; generated SDKs for Go, Python, TypeScript.
2. Single-binary mesh: Gateway (A2A JSON-RPC and gRPC), Dispatcher, embedded task store
   (SQLite or BadgerDB), no policy, no token exchange. Enough to run v1 flows on v2.
3. Agent SDK `Attach` client with credits, heartbeats, and `seq` replay.
4. Pull-mode registration against a public A2A sample agent.
5. OAuth 2.1 resource server and token exchange; SPIFFE mTLS for `Attach`.
6. Cedar policy hook; default policies; audit log.
7. OTel GenAI spans and usage roll-up.
8. MCP facade and tool-agent adapter.
9. Replicated task store (NATS JetStream) and multi-replica deployment.

## Open questions

1. **A2A bidirectional streaming.** A2A's roadmap lists it for a future release. When it lands,
   should `Attach` be re-expressed as an A2A custom binding, as AGNTCY SLIM did, and the
   bespoke `Dispatch` service deprecated? Recommendation: yes, provided the upstream design
   supports leases and credits; otherwise propose those as an A2A extension.
2. **Routing extension upstreaming.** `https://agenet.io/ext/routing/v1` (target a skill on a
   multi-agent endpoint) is generic enough to propose to the A2A project. Doing so would let
   agent gateways (Agent Router, Kong, agentgateway) front an Agenet mesh with no custom code.
3. **SLIM as the Attach transport.** AGNTCY SLIM already offers gRPC request-reply, pub/sub,
   streaming, and mandatory MLS end-to-end encryption. Adopting it would remove the need to
   specify Attach's transport but adds a dependency on a project whose AAIF status is still
   pending as of October 2026.
4. **Task store contract.** Should the spec define a storage interface (so JetStream, Kafka,
   and Postgres backends are interchangeable) or only semantics? Recommendation: semantics
   plus a conformance test suite.
5. **Cost attribute.** No stable OTel `gen_ai.cost` exists. Keep `usd` optional in
   `agenet.io/usage` until the conventions settle, and avoid inventing a competing name.
6. **MCP `subscriptions/listen` and MCP Events.** Not bridged in v2. Revisit once MCP Events
   (SEP-3415) is accepted.
7. **Browser clients.** A2A has no WebSocket binding; the Gateway could expose AG-UI events
   (`RUN_STARTED`, `TEXT_MESSAGE_CONTENT`, `STATE_DELTA`) as a thin projection of the task log
   for frontends. Out of scope for the core spec.
