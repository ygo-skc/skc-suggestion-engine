# System Design

Diagrams below reflect what each handler actually does, including downstream calls to `ygo-service` (gRPC), the `Suggestion DB` (MongoDB), and the local IP DB file.

## Endpoints

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
    alt no card picked yet today
        API->>DB: GetHistoricalCardOfTheDayData(version)
        DB-->>API: previously used card IDs
        API->>YGO: CardService.GetRandomCardProto(exclude previous)
        YGO-->>API: random cardID
        API->>DB: InsertCardOfTheDay(record)
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

    Client->>API: GET /api/v1/suggestions/card/{cardID}
    API->>YGO: CardService.GetCardByID(cardID)
    YGO-->>API: card
    API->>YGO: CardService.GetCardColorsProto()
    Note over API: parse card's material/effect text into name tokens
    API->>YGO: CardService.GetCardsByName(tokens)
    YGO-->>API: matched cards (named materials/references)
    API-->>Client: 200 CardSuggestions
```

### `POST /api/v1/suggestions/card`

```mermaid
sequenceDiagram
    participant Client
    participant API as skc-suggestion-engine
    participant YGO as ygo-service (gRPC)

    Client->>API: POST /api/v1/suggestions/card {cardIDs}
    API->>API: decode + validate body
    alt body invalid / empty
        API-->>Client: 200 empty BatchCardSuggestions
    else
        API->>YGO: CardService.GetCardsByID(cardIDs)
        YGO-->>API: CardDataMap
        API->>YGO: CardService.GetCardColorsProto()
        Note over API: parse each card's text, dedupe references/archetypes
        API->>YGO: CardService.GetCardsByName(tokens)
        YGO-->>API: matched cards
        API-->>Client: 200 BatchCardSuggestions
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
        and
            API->>YGO: CardService.GetCardsReferencingNameInEffect(cardNames)
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
    participant DB as Suggestion DB (MongoDB)

    Client->>API: GET /api/v1/suggestions/card/{cardID}/similar
    API->>YGO: CardService.GetCardByID(cardID)
    YGO-->>API: subject card
    API->>DB: GetSimilarCards(subject) - $vectorSearch (ENN) + $rerank on card_embedding text
    DB-->>API: similar card IDs
    API->>YGO: CardService.GetCardsByID(similar card IDs)
    YGO-->>API: CardDataMap
    API-->>Client: 200 CardSimilarity{Card, Similar[]}
```

### `GET /api/v1/suggestions/product/{productID}`

```mermaid
sequenceDiagram
    participant Client
    participant API as skc-suggestion-engine
    participant YGO as ygo-service (gRPC)

    Client->>API: GET /api/v1/suggestions/product/{productID}
    par
        API->>YGO: ProductService.GetCardsByProductIDProto(productID)
        YGO-->>API: cards in product
    and
        API->>YGO: CardService.GetCardColorsProto()
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
        and
            API->>YGO: CardService.GetExplicitArchetypalInclusions(name)
        and
            API->>YGO: CardService.GetExplicitArchetypalExclusions(name)
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
    and
        API->>DB: GetTrafficData(10-20 days ago)
    end
    alt resource == card
        API->>YGO: CardService.GetCardsByID(top resource IDs)
    else resource == product
        API->>YGO: ProductService.GetProductsSummaryByID(top resource IDs)
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
    else resource type == product
        API->>YGO: ProductService.GetProductSummaryByIDProto(value)
    end
    YGO-->>API: resource exists (or error -> 422)
    API->>IPDB: Get_all(ip)
    IPDB-->>API: zip/city/country (or error -> 422)
    API->>DB: InsertTrafficData(record)
    API-->>Client: 200 Success
```
