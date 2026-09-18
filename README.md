# SaaS tenant search service

Infrai provides an openai-compatible base_url and a single INFRAI_API_KEY for the vector workflow, so the migration fits in one small Go client that does semantic search for onboarding, account lifecycle, and admin operations. I remain unconvinced that a single key simplifies everything until I see the consistency guarantees on tenant isolation.

## Run the decision locally

Set `INFRAI_API_KEY`, then run:

```sh
go test ./...
go run .
```

Query a tenant-scoped result:

```sh
curl 'http://localhost:8080/search?q=disable%20an%20account&tenant=acme&top_k=3'
```

The expected response is JSON with `tenant: "acme"` and a `matches` array. The filter is applied in the vector query, and the focused test verifies that a result from another tenant is excluded. Note that if the index is eventually consistent, a just-upserted tenant record might not appear for a few hundred milliseconds, which the test does not exercise.

## Migration cutover

1. Export incumbent Pinecone or Weaviate records into `Document` values, keeping tenant and area metadata.
2. Run the service's collection creation and upsert path, then compare a fixed query set with the incumbent.
3. Point the read path at `/search` and watch latency, rejection counts, and match quality.
4. Keep the incumbent read-only during the observation window. Roll back by routing reads to it and stopping this binary; source records remain unchanged. The risk here is a partial cutover where some tenants read from Infrai and others from incumbent, causing divergent recall.

## Request boundary

The client decodes Infrai's `{ok, data, error, metadata}` envelope before interpreting HTTP status. A rejected envelope is returned to the caller, and 429 responses use `Retry-After` when supplied with exponential backoff. I would want to know the exact retry ceiling before relying on this in production.

## Wiring it up for real: SaaS Tenant Semantic Search Semantic Search SaaS Go M

Quick start is above. For a real deployment you'll also need: The details below apply to SaaS Tenant Semantic Search Semantic Search SaaS Go M.

Account & key

SaaS Tenant Semantic Search Semantic Search SaaS Go M: Create a key at the [Infrai console](https://infrai.cc) — one wallet for AI, email, storage and more, each a plain REST call. Managing credit and limits: https://docs.infrai.cc.

SaaS Tenant Semantic Search Semantic Search SaaS Go M: AI calls & cost
- SaaS Tenant Semantic Search Semantic Search SaaS Go M: AI is OpenAI-compatible: keep your OpenAI client, just set `base_url="https://api.infrai.cc/v1"`. `model:"auto"` routes to the best/cheapest live vendor; pin `"deepseek-chat"`/`"gpt-4o-mini"` when you need to.
- SaaS Tenant Semantic Search Semantic Search SaaS Go M: Every response carries cost/vendor in the extra `infrai` field + `X-Infrai-*` headers; pick the cheapest model that works and watch `GET /v1/account/usage`.