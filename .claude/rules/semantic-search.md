---
paths:
  - "db/skc_suggestion_engine_db.go"
  - "db/mongo_pipeline.go"
  - "api/card_similarity_handler.go"
  - "downstream/voyage.go"
---

# Semantic / similar-card search flow

Both `GET /card/{id}/similar` and `GET /card/search?q=` follow: **Voyage embed (query) → Mongo
`$rankFusion`** (Reciprocal Rank Fusion, k=60, weighting vector 0.65 / BM25 text 0.35; `/similar` adds
shared type/attribute/monster-type metadata boosts, `search` does not) **→ Voyage rerank → hydrate via
`ygo-service.GetCardsByID`**. The RRF constants and boost math are documented inline in `db/skc_suggestion_engine_db.go`
and `db/mongo_pipeline.go`. Embeddings are `voyage-4` @ 512 dims; reranking is `rerank-2.5`.
