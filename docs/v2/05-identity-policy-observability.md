# Agenet v2 — Identity, policy, observability

v1 had a `clientId` string in metadata and nothing else. v2 treats every hop as a separately
authenticated, authorized, and traced call, following the MCP authorization spec and the
OWASP Agentic Top 10 (ASI03 identity and privilege abuse, ASI07 insecure inter-agent
communication, ASI08 cascading failures).

## Identity

### Callers to the Gateway

- The Gateway is an OAuth 2.1 **resource server**. It publishes RFC 9728 Protected Resource
  Metadata at `/.well-known/oauth-protected-resource`, and clients MUST send an RFC 8707
  `resource` indicator so tokens are audience-bound to the mesh.
- Client registration uses Client ID Metadata Documents; Dynamic Client Registration is
  accepted only for authorization servers that lack CIMD support, mirroring MCP 2026-07-28.
- Human-initiated calls carry an end-user token. Service-initiated calls carry client
  credentials. Agents on the mesh calling back in carry a **callback token** (below); the
  Gateway never accepts a task token directly, because its audience is the agent, not the mesh.

### Task tokens: no passthrough

The Gateway never forwards the caller's token to an agent. For every delivery it mints, via
RFC 8693 token exchange against the mesh's authorization server, a **task token** with:

| Claim | Value |
|---|---|
| `aud` | the target agent's name (`agent:{tenant}/{name}`) |
| `sub` | the originating principal |
| `act` | nested actor chain: `{sub: gateway} → {sub: client} → ...` per RFC 8693 |
| `agenet_task`, `agenet_ctx` | task and context ids; the token is useless for any other task |
| `scope` | the skill id plus any downstream scopes the policy allowed |
| `exp` | the task deadline |

When the originating user's identity must reach a third-party API behind the agent, the
mesh's authorization server issues an ID-JAG assertion (IETF identity assertion authorization
grant, as shipped by Okta Cross App Access and Auth0) that the agent exchanges with the
third party's authorization server. The agent never sees a token for anything but its own
audience.

### Callback tokens: agents calling the mesh

When an agent needs another skill, it must present a credential whose audience is the mesh.
It obtains one by RFC 8693 token exchange at the mesh's authorization server:

| Parameter | Value |
|---|---|
| `subject_token` | the task token it received (`aud` = this agent) |
| `actor_token` | its workload credential (SPIFFE JWT-SVID or client-credentials token) |
| `resource` | the mesh's resource indicator |
| `scope` | the skill it intends to call |

The authorization server verifies that the actor is the task token's audience, then issues a
**callback token** with `aud` = mesh, the same `sub`, `agenet_task`, and `agenet_ctx` claims,
and the actor appended to the `act` chain. The Gateway treats it like any caller token and
performs the next exchange toward the target agent. Task-token audience isolation is
preserved: a stolen task token is still useless at the Gateway without the agent's workload
credential, and a callback token is bound to one task and context.

### Agents to the Dispatcher

- Workload identity: mTLS with SPIFFE SVIDs is the recommended default; OAuth client
  credentials in gRPC metadata are the fallback.
- The signed Agent Card binds the workload identity to the declared skills: the signing key
  must be one the tenant trusts for that `provider`.

### Delegation chain

`agenet.io/delegation` is an ordered list carried on every delivery and every child task:

```json
[
  { "kind": "user",   "sub": "alice@example.com", "iss": "https://idp.example" },
  { "kind": "client", "sub": "cli:finance-assistant" },
  { "kind": "agent",  "sub": "agent:acme/planner", "taskId": "…parent…" },
  { "kind": "agent",  "sub": "agent:acme/summarize", "taskId": "…child…" }
]
```

The chain has a maximum length (`agenet.io/budget.hops`, default 8). Policy can match on any
element, which is how a tenant says "the payments skill may only be reached when a human is
at the root of the chain."

## Policy

A policy decision point (PDP) is called twice: a cheap pre-check at the Gateway (is this
principal allowed to see and call this skill at all?) and a full check at the Dispatcher
before delivery. Engines are pluggable; Cedar, OPA/Rego, and CEL are expected. Input:

```json
{
  "tenant": "acme",
  "principal": { "sub": "alice@example.com", "iss": "https://idp.example", "roles": ["analyst"] },
  "delegation": [ ... ],
  "skill": { "id": "summarize", "version": "2.1.0", "tags": ["finance"],
             "sideEffects": "none", "dataClasses": ["pii"] },
  "agent": { "name": "summarize", "provider": "acme", "cardHash": "sha256:…", "labels": { "region": "eu" } },
  "request": { "inputModes": ["text/plain", "application/pdf"], "dataLabels": ["pii"],
               "estimatedCost": { "usd": 0.12 }, "deadlineMs": 120000 },
  "context": { "contextId": "…", "usageSoFar": { "usd": 0.80 }, "parentTaskId": "…" }
}
```

Decisions: `allow`, `deny` (task → `REJECTED` with `POLICY_DENIED`), `require_approval`
(task → `INPUT_REQUIRED` with an approval request; see [03](03-task-lifecycle.md)), and
`allow_with` (clamped budget, forced output modes, or a redaction hint for the agent). Every
decision is written to `agenet.io/policy` on the task and to the trace span.

Policies the mesh ships as defaults:

- A skill with `sideEffects: external` requires `require_approval` unless the root of the
  delegation chain is a user and the skill is explicitly allow-listed for that client.
- A request labelled `pii` or `phi` may only reach agents whose card declares the class.
- An agent may not call a skill whose provider is the agent's own provider with a longer chain
  than 2 (loop guard), unless the tenant opts out.

## Content isolation

- Routing, authorization, and scheduling read only `agenet.io/*` metadata, the token, and the
  card. They never read `parts[]`. A prompt-injected payload cannot change where a task goes.
- Agent descriptions, skill descriptions, and examples are rendered to callers as data; the
  mesh card marks them `untrusted: true` so client-side LLMs can place them accordingly.
- Agents receive the delegation chain and policy decision, so an agent that is itself an LLM
  loop can implement Meta's "Rule of Two" (do not combine untrusted input, private data, and
  external communication in one session) with mesh-provided facts rather than guesses.

## Observability

### Tracing

- `traceparent`, `tracestate`, and `baggage` are accepted from callers, propagated unchanged
  to agents, and never minted by the mesh when present. Absent trace context is created at
  the Gateway.
- Spans follow the OpenTelemetry GenAI semantic conventions (`semantic-conventions-genai`,
  v1.42.0 or later):

| Component | Span | Kind | Key attributes |
|---|---|---|---|
| Gateway | `invoke_agent {skill}` | CLIENT | `gen_ai.agent.name`, `gen_ai.conversation.id=contextId`, `agenet.task.id`, `agenet.tenant`, `agenet.policy.decision` |
| Dispatcher | `agenet.dispatch {skill}` | INTERNAL | `agenet.agent.instance`, `agenet.attempt`, `agenet.lease` |
| Agent SDK | `invoke_agent {skill}` | SERVER | `gen_ai.usage.input_tokens`, `gen_ai.usage.output_tokens`, `gen_ai.request.model` |
| Tool agent adapter | `execute_tool {tool}` | CLIENT | `gen_ai.tool.name`, `gen_ai.tool.call.id` |

### Usage and cost

Agents report `agenet.io/usage` on status events. The Dispatcher sums into the task record;
the Gateway sums into the context. `ListTasks` and `GetTask` expose both. Because no stable
`gen_ai.cost` attribute exists yet, `usd` is optional and computed by the agent or by a
mesh-side price table keyed on `model`.

### Audit log

One append-only record per task transition: `{ts, taskId, contextId, tenant, principal,
delegation, skill, agent, cardHash, fromState, toState, policy, usage}`. Exportable as OTLP
logs. This is the record a tenant produces for ASI08 "cascading failure" investigations.
