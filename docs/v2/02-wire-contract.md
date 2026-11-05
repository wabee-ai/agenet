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
   complete A2A endpoint for one registered agent, and
   `https://mesh.example/agents/{agent}/skills/{skill}` is a complete A2A endpoint for exactly
   one of its skills (its card lists only that skill). The per-skill form is the recommended
   path for third-party clients calling a multi-skill agent. On the agent-level form the
   Gateway selects the skill in this order, never reading `parts[]`:
   1. `Message.metadata["agenet.io/skill"]` if present (the key is honoured even without the
      extension header);
   2. the agent's only skill, if its card declares exactly one;
   3. otherwise the request is **agent-wide**: `Deliver` carries no skill, the agent chooses
      internally, and the Dispatcher applies the most conservative hints across the agent's
      skills (`sideEffects` = highest of `none < local < external`, `destructive` = logical
      OR, `idempotent` and `resumable` = logical AND, `dataClasses` = intersection,
      `costHint` = maximum). Policy
      sees `skill.id = "*"` with those merged values, so approval and redelivery rules can only
      be stricter than for any individual skill.
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
  uint32 attempt                = 3;            // echoed from Deliver.attempt
  uint64 seq                    = 4;            // monotonic per (task_id, attempt), starts at 1 each attempt
  a2a.v1.StreamResponse body    = 5;            // statusUpdate | artifactUpdate | message
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
    Started     started     = 16;               // Ack durably committed; safe to begin side effects
    TokenRefresh token_refresh = 17;            // fresh task token on the current lease
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
  string task_token           = 8;              // minted per attempt after instance selection; aud = this agent, bound to lease; see 05
  google.protobuf.Timestamp deadline = 9;
  map<string,string> metadata = 10;             // agenet.io/* keys and trace context
  repeated ArtifactState artifacts = 11;        // committed artifact state when resuming
  uint64 resume_offset = 12;                    // task-store offset of the last committed event
}

message ArtifactState {
  a2a.v1.Artifact artifact = 1;                 // all durably accepted parts, in order
  bool closed              = 2;                 // a chunk with lastChunk=true was committed
  uint32 chunk_count       = 3;                 // committed chunks for this artifactId
}

message Input  { string task_id = 1; string lease = 2; a2a.v1.Message message = 3;
                 string task_token = 4; }       // always a fresh token for the resumed work
message Cancel { string task_id = 1; string lease = 2; google.rpc.Status reason = 3; }
message Credit { uint32 credits = 1; }
message Shutdown { bool drain = 1; uint32 grace_seconds = 2; }
message Started { string task_id = 1; string lease = 2; uint32 attempt = 3; }
message TokenRefresh { string task_id = 1; string lease = 2; string task_token = 3; }
```

### Attach rules

- The first frame is `Register`; the Dispatcher answers `Registered` or closes the stream with
  `PERMISSION_DENIED` (bad signature, policy) or `FAILED_PRECONDITION` (card hash changed
  without a version bump).
- Delivery is credit-based. The Dispatcher never sends more `Deliver` frames than outstanding
  credits. Agents replenish with `Capacity` or by completing tasks.
- `Ack` within `ackTimeout` (default 5 s) or the lease expires and the task is redelivered.
  On `Ack` the Dispatcher durably records the lease in the task store **before** sending
  `Started`. An agent whose skill is not `idempotent` MUST NOT begin side effects until it
  receives `Started` for that `(task_id, lease, attempt)`; an idempotent skill MAY start on
  `Ack`. If the stream fails between `Ack` and `Started`, the agent treats the attempt as
  abandoned and the task store either has no committed lease (safe to redeliver) or has one
  (handled as owner loss, see [03](03-task-lifecycle.md)).
  `Heartbeat` every `heartbeatInterval` (default 10 s) listing in-flight tasks; three missed
  heartbeats release all of the instance's leases.
- **Resuming on a new instance.** Message history alone does not tell a replacement which
  output was already published, so every redelivery of a task that has emitted events
  carries the committed artifact state: `Deliver.artifacts` lists each artifact with all of its
  durably accepted parts, whether it is `closed`, and its committed chunk count, and
  `Deliver.resume_offset` is the task-store offset of the last committed event. The
  replacement MUST continue an open artifact under its existing `artifactId` with
  `append=true`, emitting only content after the committed parts, and MUST NOT re-emit
  committed parts or start a new artifact for the same output. The Dispatcher enforces this:
  an `append=true` update to a `closed` artifact, or an `append=false` update to an existing
  `artifactId`, is rejected with `FAILED_PRECONDITION` and is not written to the log. A skill
  that cannot continue from committed state must not declare itself `resumable`.
- Task tokens are short-lived and renewed on the existing lease, so a healthy task never has
  to be abandoned because its token expired. While the lease is live (heartbeats arriving),
  the Dispatcher sends `TokenRefresh` once two thirds of the current token's lifetime has
  elapsed. `Input` always carries a fresh token, which covers parked tasks whose previous
  token expired while waiting for the caller. The agent replaces its token on receipt; tokens
  are never refreshed for a lease that has been released.
- `TaskEvent.seq` is monotonic per `(task_id, attempt)` and restarts at 1 on every delivery
  attempt, so a new instance never needs to know the previous attempt's counter. The Dispatcher
  deduplicates producer events on `(task_id, attempt, seq)` and drops events whose `attempt`
  or `lease` is not the current one. The task store assigns each accepted event a separate
  task-global log offset (`agenet.io/offset`) that clients use for replay. After reconnect,
  the agent resends events the Dispatcher may have missed; duplicates are ignored.
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
  `supportedInterfaces[].protocolVersion` for A2A and `params.card.protocolVersion` on its
  routing extension entry for Agenet.
- Skill versions follow semver in the routing extension's `params.skills[<id>].version`
  (see [04](04-discovery-and-registration.md)); clients request a range.
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
