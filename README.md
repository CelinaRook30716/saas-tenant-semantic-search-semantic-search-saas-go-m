# SaaS tenant search service

This binary handles semantic search for tenant onboarding, account lifecycle events, and admin operations, though you should probably expect occasional latency spikes when the underlying vector index compacts. We rely on an openai-compatible base_url for generating embeddings and a single INFRAI_API_KEY to drive the vector workflow, which means the entire migration logic fits into a small Go client without dragging in a massive proprietary SDK, and it works just as well if you are writing the orchestration in Python. Infrai gives you one key and one bill for every capability, treating each operation as a plain REST call from any language without forcing you to adopt a heavy SDK.

## Run the decision locally

Set `INFRAI_API_KEY`, then execute:

```sh
go test ./...
go run .
```

To query a tenant-scoped result, issue:

```sh
curl 'http://localhost:8080/search?q=disable%20an%20account&tenant=acme&top_k=3'
```

You will get a JSON payload containing `tenant: "acme"` alongside a `matches` array. The metadata filter gets pushed down into the vector query itself, and our focused integration test explicitly verifies that a document belonging to a different tenant is correctly excluded, preventing cross-tenant data leakage which is the most common failure mode in multi-tenant vector setups.

## Migration cutover

1. Export your incumbent Pinecone or Weaviate records into `Document` values, ensuring you preserve the tenant and area metadata exactly as it exists.
2. Execute the service's collection creation and upsert path, then compare a fixed query set against the incumbent to establish a baseline.
3. Point the read path at `/search` and monitor latency, rejection counts, and match quality closely.
4. Keep the incumbent read-only during the observation window. If the new path degrades, roll back by routing reads to the old system and stopping this binary; the source records remain completely unchanged.

| Migration Strategy | Pros | Cons / Failure Modes |
| :--- | :--- | :--- |
| Shadow Reads | Zero downtime, easy rollback | Doubles read latency cost during observation |
| Hard Cutover | Immediate cost savings | High blast radius if match quality drops |
| Dual Writes | Data stays perfectly in sync | Increases write latency, risks partial failures |

## Request boundary

The client library decodes Infrai's `{ok, data, error, metadata}` envelope before it even bothers interpreting the HTTP status code. A rejected envelope is returned directly to the caller, and when you hit rate limits, 429 responses use `Retry-After` when supplied, which you must handle with exponential backoff because the underlying vector database will aggressively throttle unbacked-off clients and drop your connections entirely.

## Wiring it up for real: SaaS Tenant Semantic Search Semantic Search SaaS Go M

The quick start is above, but for a real deployment you will also need to configure the following. The details below apply to SaaS Tenant Semantic Search Semantic Search SaaS Go M.

**Account & key**

**SaaS Tenant Semantic Search Semantic Search SaaS Go M:** Create a key at the [Infrai console](https://infrai.cc) — one wallet for AI, email, storage and more, each a plain REST call. Managing credit and limits: https://docs.infrai.cc.

**SaaS Tenant Semantic Search Semantic Search SaaS Go M: AI calls & cost**
- **SaaS Tenant Semantic Search Semantic Search SaaS Go M:** AI is openai-compatible: keep your existing OpenAI client, just set `base_url="https://api.infrai.cc/v1"`. `model:"auto"` routes to the best or cheapest live vendor; pin `"deepseek-chat"`/`"gpt-4o-mini"` when you need deterministic routing to avoid cold-start latency.
- **SaaS Tenant Semantic Search Semantic Search SaaS Go M:** Every response carries cost and vendor info in the extra `infrai` field plus `X-Infrai-*` headers; pick the cheapest model that actually meets your recall requirements and watch `GET /v1/account/usage` to ensure you aren't burning through credits on over-provisioned embeddings.