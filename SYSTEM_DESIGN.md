# System Design

Diagrams below reflect what each handler actually does, including downstream calls to `ygo-service` (gRPC), the `Suggestion DB` (MongoDB), and the local IP DB file.

## Endpoints (v1)

### `GET /api/v1/suggestions/status`

```mermaid
sequenceDiagram
    participant Client
    participant API as skc-suggestion-engine
    participant YGO as ygo-service (gRPC)
    participant DB as Suggestion DB (MongoDB)

    Client->>API: GET /api/v1/suggestions/status
    API->>YGO: HealthService.GetAPIStatus()
    YGO-->>API: version (or error -> Down)
    API->>DB: GetSKCSuggestionDBVersion()
    DB-->>API: version (or error -> Down)
    API-->>Client: 200 APIHealth{version, downstream[]}
```

### `POST /api/v1/suggestions/card-details`

```mermaid
sequenceDiagram
    participant Client
    participant API as skc-suggestion-engine
    participant YGO as ygo-service (gRPC)

    Client->>API: POST /api/v1/suggestions/card-details {cardIDs}
    API->>API: decode + validate body
    alt body invalid / empty
        API-->>Client: 200 empty BatchCardData
    else
        API->>YGO: CardService.GetCardsByID(cardIDs)
        YGO-->>API: CardDataMap + UnknownResources
        API-->>Client: 200 BatchCardData
    end
```

### `GET /api/v1/suggestions/card-of-the-day`

```mermaid
sequenceDiagram
    participant Client
    participant API as skc-suggestion-engine
    participant DB as Suggestion DB (MongoDB)
    participant YGO as ygo-service (gRPC)

    Client->>API: GET /api/v1/suggestions/card-of-the-day
    API->>DB: GetCardOfTheDay(today, version)
    DB-->>API: existing cardID (or none)
    alt no card picked yet today
        API->>DB: GetHistoricalCardOfTheDayData(version)
        DB-->>API: previously used card IDs
        API->>YGO: CardService.GetRandomCardProto(exclude previous)
        YGO-->>API: random cardID
        API->>DB: InsertCardOfTheDay(record)
        DB-->>API: ack
    end
    API->>YGO: CardService.GetCardByID(cardID)
    YGO-->>API: card details
    API-->>Client: 200 CardOfTheDay
```

### `GET /api/v1/suggestions/card/{cardID}`

```mermaid
sequenceDiagram
    participant Client
    participant API as skc-suggestion-engine
    participant YGO as ygo-service (gRPC)
    participant DB as Suggestion DB (MongoDB)

    Client->>API: GET /api/v1/suggestions/card/{cardID}
    API->>YGO: CardService.GetCardByID(cardID)
    YGO-->>API: card
    par suggest.FetchMetadata
        API->>YGO: CardService.GetCardColorsProto()
        YGO-->>API: card color IDs
    and
        API->>DB: GetRelevantArchetypes([cardID])
        DB-->>API: relevant archetypes
    end
    Note over API: parse card's material/effect text into name tokens<br/>(archetypes from DB seed the archetype set)
    API->>YGO: CardService.GetCardsByName(tokens)
    YGO-->>API: matched cards (named materials/references)
    API-->>Client: 200 CardSuggestions{..., relevantArchetypes}
```

### `POST /api/v1/suggestions/card`

```mermaid
sequenceDiagram
    participant Client
    participant API as skc-suggestion-engine
    participant YGO as ygo-service (gRPC)
    participant DB as Suggestion DB (MongoDB)

    Client->>API: POST /api/v1/suggestions/card {cardIDs}
    API->>API: decode + validate body
    alt body invalid / empty
        API-->>Client: 200 empty BatchCardSuggestions
    else
        API->>YGO: CardService.GetCardsByID(cardIDs)
        YGO-->>API: CardDataMap
        par suggest.FetchMetadata
            API->>YGO: CardService.GetCardColorsProto()
            YGO-->>API: card color IDs
        and
            API->>DB: GetRelevantArchetypes(cardIDs)
            DB-->>API: relevant archetypes
        end
        Note over API: parse each card's text, dedupe references/archetypes<br/>(archetypes from DB seed the archetype set)
        API->>YGO: CardService.GetCardsByName(tokens)
        YGO-->>API: matched cards
        API-->>Client: 200 BatchCardSuggestions{..., relevantArchetypes}
    end
```

### `GET /api/v1/suggestions/card/support/{cardID}`

```mermaid
sequenceDiagram
    participant Client
    participant API as skc-suggestion-engine
    participant YGO as ygo-service (gRPC)

    Client->>API: GET /api/v1/suggestions/card/support/{cardID}
    API->>YGO: CardService.GetCardByID(cardID)
    YGO-->>API: card
    API->>YGO: CardService.GetCardsReferencingNameInEffect([cardName])
    YGO-->>API: candidate referencing cards
    Note over API: split into "material for" vs "referenced by"
    API-->>Client: 200 CardSupport
```

### `POST /api/v1/suggestions/card/support`

```mermaid
sequenceDiagram
    participant Client
    participant API as skc-suggestion-engine
    participant YGO as ygo-service (gRPC)

    Client->>API: POST /api/v1/suggestions/card/support {cardIDs}
    API->>API: decode + validate body
    alt body invalid / empty
        API-->>Client: 200 empty BatchCardSupport
    else
        API->>YGO: CardService.GetCardsByID(cardIDs)
        YGO-->>API: CardDataMap
        par
            API->>YGO: CardService.GetCardColorsProto()
            YGO-->>API: card color IDs
        and
            API->>YGO: CardService.GetCardsReferencingNameInEffect(cardNames)
            YGO-->>API: candidate referencing cards
        end
        Note over API: split into "material for" vs "referenced by" per card
        API-->>Client: 200 BatchCardSupport
    end
```

### `GET /api/v1/suggestions/card/{cardID}/similar`

```mermaid
sequenceDiagram
    participant Client
    participant API as skc-suggestion-engine
    participant YGO as ygo-service (gRPC)
    participant Voyage as Voyage AI
    participant DB as Suggestion DB (MongoDB)

    Client->>API: GET /api/v1/suggestions/card/{cardID}/similar
    API->>YGO: CardService.GetCardByID(cardID)
    YGO-->>API: subject card
    API->>Voyage: EmbedText(subject effect, input_type=query)
    Voyage-->>API: query embedding
    API->>DB: VectorSearchOnCardEmbedding(subject, embedding) - $vectorSearch (ENN, limit 30)
    DB-->>API: candidate results (id + text)
    API->>Voyage: RerankVectorResults(candidate texts, subject effect, topK=20)
    Voyage-->>API: reranked results
    API->>YGO: CardService.GetCardsByID(reranked card IDs)
    YGO-->>API: CardDataMap
    API-->>Client: 200 SimilarCards{Card, Matches[]}
```

### `GET /api/v1/suggestions/product/{productID}`

```mermaid
sequenceDiagram
    participant Client
    participant API as skc-suggestion-engine
    participant YGO as ygo-service (gRPC)
    participant DB as Suggestion DB (MongoDB)

    Client->>API: GET /api/v1/suggestions/product/{productID}
    API->>YGO: ProductService.GetCardsByProductIDProto(productID)
    YGO-->>API: cards in product
    par suggest.FetchMetadata (cardIDs from product)
        API->>YGO: CardService.GetCardColorsProto()
        YGO-->>API: card color IDs
    and
        API->>DB: GetRelevantArchetypes(cardIDs)
        DB-->>API: relevant archetypes
    end
    par
        Note over API: getBatchSuggestions (may call GetCardsByName)
    and
        Note over API: getBatchSupport (calls GetCardsReferencingNameInEffect)
    end
    API-->>Client: 200 ProductSuggestions{Suggestions, Support}
```

### `GET /api/v1/suggestions/archetype/{archetypeName}`

```mermaid
sequenceDiagram
    participant Client
    participant API as skc-suggestion-engine
    participant DB as Suggestion DB (MongoDB)
    participant YGO as ygo-service (gRPC)

    Client->>API: GET /api/v1/suggestions/archetype/{archetypeName}
    API->>API: validate archetype name format
    API->>DB: IsBlackListed("archetype", name)
    alt blacklisted
        API-->>Client: 422 blacklisted archetype
    else
        par
            API->>YGO: CardService.GetArchetypalCardsUsingCardName(name)
            YGO-->>API: cards matching name
        and
            API->>YGO: CardService.GetExplicitArchetypalInclusions(name)
            YGO-->>API: explicitly included cards
        and
            API->>YGO: CardService.GetExplicitArchetypalExclusions(name)
            YGO-->>API: explicitly excluded cards
        end
        alt fewer than 2 cards found by name
            API-->>Client: 404 not an archetype
        else
            Note over API: remove excluded cards from "using name" results
            API-->>Client: 200 ArchetypalSuggestions
        end
    end
```

### `GET /api/v1/suggestions/trending/{resource}`

```mermaid
sequenceDiagram
    participant Client
    participant API as skc-suggestion-engine
    participant DB as Suggestion DB (MongoDB)
    participant YGO as ygo-service (gRPC)

    Client->>API: GET /api/v1/suggestions/trending/{card|product}
    par
        API->>DB: GetTrafficData(last 10 days)
        DB-->>API: recent traffic records
    and
        API->>DB: GetTrafficData(10-20 days ago)
        DB-->>API: previous traffic records
    end
    alt resource == card
        API->>YGO: CardService.GetCardsByID(top resource IDs)
        YGO-->>API: CardDataMap
    else resource == product
        API->>YGO: ProductService.GetProductsSummaryByID(top resource IDs)
        YGO-->>API: product summaries
    end
    Note over API: compute occurrence + rank change vs. previous period
    API-->>Client: 200 Trending{metrics}
```

### `POST /api/v1/suggestions/traffic-analysis` 🔒 (requires `API-Key` header)

```mermaid
sequenceDiagram
    participant Client
    participant API as skc-suggestion-engine
    participant YGO as ygo-service (gRPC)
    participant IPDB as IP DB (local file)
    participant DB as Suggestion DB (MongoDB)

    Client->>API: POST /api/v1/suggestions/traffic-analysis {resource, ip, source}
    API->>API: verifyAPIKeyMiddleware (checks API-Key header)
    alt invalid/missing key
        API-->>Client: 401 Unauthorized
    end
    API->>API: decode + validate body
    alt resource type == card
        API->>YGO: CardService.GetCardByID(value)
        YGO-->>API: card (or error -> 422)
    else resource type == product
        API->>YGO: ProductService.GetProductSummaryByIDProto(value)
        YGO-->>API: product summary (or error -> 422)
    end
    API->>IPDB: Get_all(ip)
    IPDB-->>API: zip/city/country (or error -> 422)
    API->>DB: InsertTrafficData(record)
    DB-->>API: ack
    API-->>Client: 200 Success
```

## Endpoints (v2)

### `GET /api/v2/suggestions/archetype/{archetypeName}`

```mermaid
sequenceDiagram
    participant Client
    participant API as skc-suggestion-engine
    participant DB as Suggestion DB (MongoDB)
    participant YGO as ygo-service (gRPC)

    Client->>API: GET /api/v2/suggestions/archetype/{archetypeName}
    API->>API: validate archetype name format
    API->>DB: GetArchetypeMembers(name)
    DB-->>API: inheritMembers, qualifiedMembers, excludedMembers (or 404 not found)
    API->>YGO: CardService.GetCardsByID(inherit + qualified + excluded IDs)
    YGO-->>API: CardDataMap
    Note over API: sort each member list by card name
    API-->>Client: 200 ArchetypeMembers{InheritMembers, QualifiedMembers, ExcludedMembers}
```

Unlike the v1 `/archetype/{archetypeName}` endpoint (which derives membership by scanning card names/text via `ygo-service` and has no explicit-exclusion source of truth beyond that scan), v2 reads a curated membership document straight from the Suggestion DB (`inheritMembers`, `qualifiedMembers`, `excludedMembers` fields) and hydrates it with card data.
