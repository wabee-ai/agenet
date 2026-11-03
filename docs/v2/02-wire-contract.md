# Agenet v2 — Wire contract

The protobuf schema is the source of truth. JSON-RPC and HTTP+JSON projections follow A2A's
binding rules so that any A2A SDK can talk to the Gateway without modification.

## Client side: unmodified A2A

The Gateway implements the A2A 1.0 `A2AService` exactly as specified: `SendMessage`,
`SendStreamingMessage`, `GetTask`, `ListTasks`, `CancelTask`, `SubscribeToTask`,
`CreateTaskPushNotificationConfig` and its Get/List/Delete siblings, `GetAgentCard`, and
`GetExtendedAgentCard`. Errors use `google.rpc.Status` with A2A's `ErrorInfo` reasons.

Two ways to target a skill:

1. **Virtual endpoint** (no extension needed). `https://mesh.example/agents/{agent}` is a
   complete A2A endpoint for one registered agent. Its card lists that agent's skills. This is
   the recommended path for third-party clients.
2. **Mesh endpoint with the routing extension.** `https://mesh.example/` accepts any skill on
   the mesh. The client sends the `A2A-Extensions: https://agenet.io/ext/routing/v1` header and
   sets routing keys in `Message.metadata`:

```json
{
  "message": {
    "messageId": "0193e8b2-7c1a-7f1e-9a4e-2d6f0c1a9b77",
    "role": "ROLE_USER",
    "parts": [{ "text": "Summarize the attached report", "mediaType": "text/plain" },
              { "url": "https://files.example/q3.pdf", "mediaType": "application/pdf" }],
    "metadata": {
      "agenet.io/skill": "summarize",
      "agenet.io/skillVersion": ">=2 <3",
      "agenet.io/tags": ["finance"],
      "agenet.io/deadlineMs": 120000,
      "agenet.io/priority": 5,
      "agenet.io/budget": { "inputTokens": 200000, "outputTokens": 20000, "usd": 1.50, "hops": 4 }
    }
  },
  "configuration": { "acceptedOutputModes": ["text/markdown"], "returnImmediately": false }
}
```

`agenet.io/skill` is the direct heir of v1's `requestType`. Everything under `agenet.io/` is
reserved; the Gateway rejects unknown keys in that namespace with `INVALID_ARGUMENT`.

## Reserved metadata keys

Keys are mirrored into gRPC metadata (`agenet-skill`, `agenet-tenant`, ...) and HTTP headers
(`Agenet-Skill`, ...) so that L7 routers can act on them without parsing the body. The receiving
component MUST verify header and body agree and MUST reject on mismatch, following MCP's
`Mcp-Method`/`HeaderMismatch` rule.

| Key | Set by | Meaning |
|---|---|---|
| `agenet.io/skill`, `skillVersion`, `tags`, `agent` | client | Routing key. `agent` pins a specific registered agent name. |
| `agenet.io/tenant` | Gateway (from token) | Multi-tenancy partition; also fills A2A's `tenant` field on gRPC |
| `agenet.io/principal` | Gateway | Stable identifier of the authenticated caller (`sub@iss`) |
| `agenet.io/delegation` | Gateway | Ordered list of `{sub, iss, kind: user|client|agent, taskId?}`; see [05](05-identity-policy-observability.md) |
| `agenet.io/deadlineMs`, `priority`, `budget` | client, clamped by policy | Scheduling and cost limits |
| `agenet.io/attempt`, `lease` | Dispatcher | Delivery attempt counter and lease id, visible to the agent |
| `agenet.io/usage` | agent, on status events | `{inputTokens, outputTokens, cacheReadTokens, reasoningTokens, model, usd?}` |
| `agenet.io/policy` | Dispatcher | `{decision, policyId, reason}` recorded on the task |
| `traceparent`, `tracestate`, `baggage` | caller, propagated | W3C trace context, same keys MCP uses in `_meta` |

`messageId` is the idempotency key (see [03](03-task-lifecycle.md)). `contextId` is the
conversation key and is server-generated on the first message of a conversation, per A2A.

## Agent side: the `Attach` stream

The one place Agenet defines its own RPCs. An agent opens a single long-lived bidirectional
stream to a Dispatcher. Correctness never depends on the stream: every frame carries the ids
needed to reconcile after a reconnect, and the task store is authoritative.

```proto
syntax = "proto3";
package agenet.v2;

import "a2a/v1/a2a.proto";           // AgentCard, Message, Task, StreamResponse, ...
import "google/protobuf/timestamp.proto";
import "google/rpc/status.proto";

service Dispatch {
  // Agent dials out. First AgentFrame MUST be Register.
  rpc Attach(stream AgentFrame) returns (stream DispatchFrame);
}

message AgentFrame {
  string frame_id = 1;                           // for acks of control frames
  oneof kind {
    Register       register        = 10;
    Heartbeat      heartbeat       = 11;
    Capacity       capacity        = 12;
    Ack            ack             = 13;         // task received; starts the lease clock
    Reject         reject          = 14;         // agent refuses before starting; redeliverable
    TaskEvent      event           = 15;         // wraps a2a.v1.StreamResponse
    Detach         detach          = 16;         // graceful drain
  }
}

message Register {
  a2a.v1.AgentCard card         = 1;            // MUST be signed (card.signatures)
  string instance_id            = 2;            // unique per process
  string protocol_version       = 3;            // "2.0"
  repeated string extensions    = 4;            // agenet.io/ext/* this instance supports
  Capacity capacity             = 5;
  map<string,string> labels     = 6;            // region, version, canary, ...
}

message Capacity { uint32 max_inflight = 1; uint32 credits = 2; }   // credit-based backpressure
message Heartbeat { google.protobuf.Timestamp at = 1; repeated string inflight_task_ids = 2; }
message Ack       { string task_id = 1; string lease = 2; }
message Reject    { string task_id = 1; string lease = 2; google.rpc.Status reason = 3; }
message Detach    { bool drain = 1; }

message TaskEvent {
  string task_id                = 1;
  string lease                  = 2;
  uint64 seq                    = 3;            // monotonic per task; Dispatcher dedupes on (task_id, seq)
  a2a.v1.StreamResponse body    = 4;            // statusUpdate | artifactUpdate | message
}

message DispatchFrame {
  string frame_id = 1;
  oneof kind {
    Registered  registered  = 10;
    Deliver     deliver     = 11;
    Input       input       = 12;               // follow-up message on a parked task
    Cancel      cancel      = 13;
    Credit      credit      = 14;               // Dispatcher grants more credits
    Shutdown    shutdown    = 15;               // Dispatcher is going away; reconnect elsewhere
  }
}

message Registered {
  string agent_url            = 1;              // virtual endpoint assigned on the Gateway
  string card_hash            = 2;              // pinned; changing the card requires re-Register
  google.protobuf.Timestamp lease_ttl = 3;
  repeated string dispatcher_endpoints = 4;     // for reconnect
}

message Deliver {
  string task_id              = 1;
  string context_id           = 2;
  string lease                = 3;
  uint32 attempt              = 4;
  a2a.v1.Message message      = 5;
  a2a.v1.SendMessageConfiguration configuration = 6;
  repeated a2a.v1.Message history = 7;          // when resuming on a new instance
  string task_token           = 8;              // audience-bound, task-scoped; see 05
  google.protobuf.Timestamp deadline = 9;
  map<string,string> metadata = 10;             // agenet.io/* keys and trace context
}

message Input  { string task_id = 1; string lease = 2; a2a.v1.Message message = 3; }
message Cancel { string task_id = 1; string lease = 2; google.rpc.Status reason = 3; }
message Credit { uint32 credits = 1; }
message Shutdown { bool drain = 1; uint32 grace_seconds = 2; }
```

### Attach rules

- The first frame is `Register`; the Dispatcher answers `Registered` or closes the stream with
  `PERMISSION_DENIED` (bad signature, policy) or `FAILED_PRECONDITION` (card hash changed
  without a version bump).
- Delivery is credit-based. The Dispatcher never sends more `Deliver` frames than outstanding
  credits. Agents replenish with `Capacity` or by completing tasks.
- `Ack` within `ackTimeout` (default 5 s) or the lease expires and the task is redelivered.
  `Heartbeat` every `heartbeatInterval` (default 10 s) listing in-flight tasks; three missed
  heartbeats release all of the instance's leases.
- `TaskEvent.seq` is per-task and monotonic. After reconnect, the agent resends events the
  Dispatcher may have missed; the Dispatcher ignores duplicates.
- `Shutdown{drain=true}` means finish in-flight tasks, send no new ones, then reconnect to one
  of `dispatcher_endpoints`.

### Transport bindings for Attach

| Binding | Framing | Notes |
|---|---|---|
| gRPC (primary) | protobuf | HTTP/2 bidi stream; mTLS with SPIFFE SVIDs recommended |
| WebSocket | protobuf or JSON (`application/agenet+json`) | For runtimes without gRPC (browsers, edge workers) |

A2A's roadmap lists bidirectional streaming for a future release. When it ships, Agenet SHOULD
offer `Attach` as an A2A custom binding, the same way AGNTCY SLIM does, and deprecate the
bespoke service per the lifecycle policy.

## Versioning

- Protocol version `2.0`. Clients send `A2A-Version` as A2A requires; the Gateway's card lists
  `supportedInterfaces[].protocolVersion` for A2A and `agenet.io/version` in `capabilities.extensions`.
- Skill versions follow semver in `AgentSkill.id` suffix (`summarize@2.1.0`) or in the skill's
  `metadata["agenet.io/version"]`; clients request a range.
- Features move Active → Deprecated (≥ 12 months, announced in the mesh card) → Removed.

## Error mapping

| Situation | Status | A2A `ErrorInfo.reason` |
|---|---|---|
| Unknown or unauthorized skill | `NOT_FOUND` | `SKILL_NOT_FOUND` |
| No instance available before deadline | `UNAVAILABLE` | `NO_CAPACITY` (task → `REJECTED`) |
| Policy denied | `PERMISSION_DENIED` | `POLICY_DENIED` with `policyId` |
| Header/body mismatch | `INVALID_ARGUMENT` | `HEADER_MISMATCH` |
| Duplicate `messageId` with different payload | `ALREADY_EXISTS` | `IDEMPOTENCY_CONFLICT` |
| Budget exhausted mid-task | task → `FAILED` | `BUDGET_EXCEEDED` |
| Task not cancelable, not found, unsupported op | per A2A | `TaskNotCancelable`, `TaskNotFound`, `UnsupportedOperation` |
