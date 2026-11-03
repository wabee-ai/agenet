# Agenet v2 — Overview

**Status:** Draft, November 2026. Supersedes the v1 documents under `docs/overview`,
`docs/architecture`, `docs/components`, and `docs/implementation`, which are kept for history.

## One-paragraph definition

Agenet v2 is an **agent mesh**: a brokered transport binding and routing fabric for the
[A2A 1.0](https://a2a-protocol.org/latest/specification/) protocol. Agents join the mesh by
dialing out to a Dispatcher and presenting a signed Agent Card. Clients reach any skill on the
mesh through one Gateway endpoint that speaks standard A2A (JSON-RPC, gRPC, or HTTP+JSON) and,
optionally, MCP. Between them sits a durable task store, so no request's correctness depends
on a live connection. Identity, policy, tracing, and token accounting are carried on every
hop as structured metadata rather than left to each agent.

Agenet v2 does **not** define a new payload format, task model, or discovery document. It
imports A2A's `Message`, `Part`, `Artifact`, `Task`, `TaskState`, and `AgentCard` unchanged,
and adds only what A2A leaves to infrastructure.

## Why this shape

Three findings from the [landscape review](../research/landscape-2024-2026.md) drive the design:

1. **A2A 1.0 and MCP 2026-07-28 are the contracts.** Both live under the Agentic AI
   Foundation. Every major cloud ships them. A new wire format would be adopted by no one.
2. **Stateful bidirectional sessions lost.** MCP ran sessions from March 2025 and removed them
   in July 2026 because they fought load balancers, enabled hijacking, and made resumption
   fragile. A2A never had them. v1's "persistent response channel" is the same mistake.
3. **Dial-out registration is still unsolved upstream.** A2A assumes an agent is an HTTP
   server with a URL. Agents on laptops, in private VPCs, or at the edge cannot be reached that
   way. v1's Dispatcher, where agents connect outward, is the one v1 idea worth keeping and
   is the mesh's differentiator. AGNTCY SLIM is the only comparable effort and it, too, is
   positioned as an A2A binding rather than a competitor.

## Design principles

| # | Principle | Consequence |
|---|---|---|
| P1 | Borrow, don't invent | Payload, task state, card, and errors are A2A's. Agenet adds metadata keys under a reserved `agenet.io/` prefix and one A2A extension URI. |
| P2 | Stateless requests, durable tasks | Every request is self-describing (version, tenant, principal, trace, idempotency key). Streams may drop at any time; the task store is the source of truth. |
| P3 | Route on declared skills | Routing key is `AgentCard.skills[].id` plus version and tags. Routing never reads the payload. |
| P4 | Identity on every hop | Audience-bound token per hop, no token passthrough, delegation chain carried, Agent Cards signed and pinned by hash. |
| P5 | Policy in the path | The Dispatcher calls a policy decision point with structured attributes before delivery; the decision is recorded in the trace. |
| P6 | Observable by default | `traceparent` propagated, OTel GenAI spans at Gateway and agent, usage and cost rolled up per task and per context. |
| P7 | Payload is untrusted | Control fields live in gRPC metadata or the `agenet.io/*` metadata map and are validated against the body; descriptions and parts are never executed or routed on. |
| P8 | Versioned, with a lifecycle | Features move Active → Deprecated (≥12 months) → Removed, matching MCP's policy. Agents advertise supported versions per interface, matching A2A. |

## What changed from v1, at a glance

| v1 concept | v2 replacement |
|---|---|
| `requestType` | `skill` (A2A `AgentSkill.id`), optionally `@version`, with tags |
| `uuid` | `taskId` (work), `contextId` (conversation), `messageId` (idempotency) |
| `data` + `content-type` | `Message.parts[]` / `Artifact.parts[]` with per-part `mediaType` |
| "message delivered" / partial / "completed" / "error" | `TaskStatusUpdateEvent` and `TaskArtifactUpdateEvent`; terminal states `COMPLETED`, `FAILED`, `CANCELED`, `REJECTED` |
| Request Channel + Response UUID Channel | Durable task store with per-skill work queues and a per-task append-only event log |
| Gateway bidi stream with client | Standard A2A operations in any binding; streaming is optional and resumable via `SubscribeToTask` |
| Dispatcher bidi stream with agent | `Attach` stream with registration, leases, credit-based backpressure, and redelivery |
| `metadata.clientId` | Authenticated principal, tenant, delegation chain |
| — | Policy decision point, trace propagation, usage accounting, MCP bridge |

## Document map

- [01 Architecture](01-architecture.md): components, flows, diagrams
- [02 Wire contract](02-wire-contract.md): proto sketch, metadata keys, bindings
- [03 Task lifecycle](03-task-lifecycle.md): states, delivery guarantees, cancel, pause, idempotency
- [04 Discovery and registration](04-discovery-and-registration.md): Agent Cards, dial-out, pull mode, virtual endpoints
- [05 Identity, policy, observability](05-identity-policy-observability.md)
- [06 MCP bridge](06-mcp-bridge.md): MCP servers as agents, agents as MCP tools
- [07 Migration and open questions](07-migration-and-open-questions.md)

## Non-goals

- Not an orchestration framework. Planning, handoffs, and loops belong in the agents
  (OpenAI Agents SDK, Microsoft Agent Framework, Claude Agent SDK, LangGraph) that sit on the mesh.
- Not an LLM gateway. Model routing and prompt caching are a separate concern; the mesh only
  carries their usage numbers.
- Not a registry of record. The mesh keeps a live view of attached agents; a durable catalog
  (MCP Registry, AGNTCY Directory) can be fed from it.
