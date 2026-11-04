# Agenet v2 — MCP bridge

MCP is the vertical protocol: agent → tools and data. Agenet speaks it in both directions so
that tools join the mesh as skills and mesh skills appear as tools to any MCP host. The
bridge targets MCP revision 2026-07-28 (stateless core, Tasks extension, MRTR).

## Direction 1: MCP servers as agents ("tool agents")

The tool-agent adapter is a small process that:

1. Connects to an MCP server (stdio or Streamable HTTP), calls `server/discover` then
   `tools/list`, and pins the result by hash. A changed tool list or description triggers
   re-registration under the same rules as any card change.
2. Synthesizes an `AgentCard` with one `AgentSkill` per tool:
   - `id` = tool `name`; `description`, `title` copied; `inputModes=["application/json"]`,
     `outputModes=["application/json", "text/plain"]`.
   - Routing extension `params.skills[<tool>].inputSchema` = `inputSchema`; `outputSchema`
     likewise (A2A skills carry no `metadata`; see [04](04-discovery-and-registration.md)).
   - `params.skills[<tool>].sideEffects` derived from annotations **after applying MCP's defaults**
     (`readOnlyHint` defaults to `false`, `destructiveHint` to `true`, `idempotentHint` to
     `false`, `openWorldHint` to `true`): `readOnlyHint=true` → `none`; otherwise
     `destructiveHint=true` → `external`; otherwise `local`. An unannotated tool is therefore
     `external` and falls under the default approval policy. `params.skills[<tool>].idempotent`
     = `idempotentHint` after defaults.
3. Registers in dial-out mode and, on `Deliver`, translates:
   - The first `data` part of the message → `tools/call` arguments, validated against the schema.
   - `structuredContent` → a `data` artifact; `content[]` → text/raw/url artifacts;
     `isError` → `FAILED` with the content as the status message.
   - A `tools/call` that returns a task handle (Tasks extension) → the Agenet task stays
     `WORKING`; the adapter polls `tasks/get` and forwards progress.
   - An MRTR `resultType: "input_required"` → `INPUT_REQUIRED` with the `inputRequests`
     as a data part; the caller's follow-up is sent back as `inputResponses` with the echoed
     `requestState`, which the adapter stores on the task.
4. Propagates `traceparent` into `_meta` and the task token as the MCP bearer token after
   exchange for the MCP server's audience.

Resources and prompts are exposed as read-only skills `resources/read` and `prompts/get`
with the URI or name as the argument; `subscriptions/listen` is not bridged in v2.

## Direction 2: mesh skills as MCP tools (the Gateway's MCP facade)

The Gateway exposes `https://mesh.example/mcp` as a stateless MCP server:

| MCP method | Agenet behaviour |
|---|---|
| `server/discover` | Returns protocol versions, `capabilities.tools`, and the Tasks extension |
| `tools/list` | One tool per visible skill; `inputSchema` from the routing extension's `params.skills[<id>].inputSchema` or a generic `{parts[]}` schema; annotations derived from `sideEffects` and `idempotent`; deterministic order; `ttlMs`/`cacheScope` from the card index |
| `tools/call` | `SendMessage` with `returnImmediately=false`. The facade reads the per-request client capabilities in `_meta["io.modelcontextprotocol/clientCapabilities"]`. **If the client advertises the Tasks extension**: results that complete within `agenet.io/syncWaitMs` (default 10 s) are returned inline as `structuredContent`; otherwise a `CreateTaskResult` handle is returned. **If it does not**: the facade never changes response shape; it blocks up to `agenet.io/syncMaxWaitMs` (default 300 s, clamped by the task deadline), streaming `notifications/progress` on the response stream, then returns inline on completion or a tool result with `isError=true` and reason `TASK_TIMEOUT` carrying the Agenet `taskId` so the caller can retrieve it via A2A `GetTask` |
| `tasks/get`, `tasks/cancel` | `GetTask`, `CancelTask` |
| `tasks/update` | `SendMessage` follow-up on a parked task |
| `INPUT_REQUIRED` while synchronous | MRTR `resultType: "input_required"`; `requestState` is the task id sealed with the mesh's AEAD key, bound to the principal and task, and expiring at the deadline |

Headers: the facade requires `Mcp-Method` and `Mcp-Name`, mirrors `agenet.io/skill` into
`Mcp-Param-Skill` via the `x-mcp-header` annotation, and validates header against body. Auth
is the same OAuth 2.1 resource server as the A2A ingress, so one token works on both facades.

Deprecated MCP features (roots, sampling, logging) are not implemented. An MCP host that wants
an agent to call back into a model does so through its own provider API, as the MCP spec now
recommends.

## MCP Apps and Skills over MCP

A tool agent may forward `ui://` resources (MCP Apps) and `skills/*` resources (Skills over
MCP) as `url` parts with their media types. The mesh does not interpret them.
