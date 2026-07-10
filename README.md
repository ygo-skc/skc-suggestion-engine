# skc-suggestion-engine

[![Unit Test](https://github.com/ygo-skc/skc-suggestion-engine/actions/workflows/unit-test.yaml/badge.svg?branch=release)](https://github.com/ygo-skc/skc-suggestion-engine/actions/workflows/unit-test.yaml) [![CodeQL](https://github.com/ygo-skc/skc-suggestion-engine/actions/workflows/codeql.yml/badge.svg?branch=release)](https://github.com/ygo-skc/skc-suggestion-engine/actions/workflows/codeql.yml)

## Info

Go API that extends functionality of [SKC API](https://github.com/ygo-skc/skc-api) with the following:

* Suggest materials and other named references by parsing the text of a card, individually or in batch
* Suggest support cards for a given card or batch of cards by analyzing every card in the DB
* Suggest related cards for a product or batch of product
* Suggest cards belonging to an archetype
* Card of the Day - a card is chosen and cached daily
* Track and report trending cards/products based on submitted traffic data
* Clients can send browsing/traffic data to build the suggestion and trending database.
* Status endpoint that reports health of the API and its downstream dependencies (SKC DB, Suggestion DB)

## Languages & Tools

* **Go** - primary language the API is written in
* **Bash**
* **YAML**
* **[chi](https://github.com/go-chi/chi)** - HTTP router/middleware
* **gRPC / Protocol Buffers** - communication with the downstream `ygo-service`
* **MongoDB** (`mongo-driver`) - persistence for suggestion/traffic data
* **[go-playground/validator](https://github.com/go-playground/validator)** - request payload validation
* **[rs/cors](https://github.com/rs/cors)** - CORS handling
* **[ip2location](https://github.com/ip2location/ip2location-go)** - IP geolocation used for traffic/trending analysis
* **[testify](https://github.com/stretchr/testify)** - unit testing
* **TLS 1.3 / HTTP2** - the API is served exclusively over HTTPS
* **Docker / Docker Compose** - containerized local and prod runtime
* **GitHub Actions** - CI for unit tests and CodeQL scanning
* **CodeQL** - static analysis/security scanning
* **AWS Secrets Manager** (via AWS CLI + `jq`) - secret/config retrieval for local setup

## Local Setup

In order for the API to work locally, do the following steps

1. Run `go mod tidy` to download deps
2. Execute the shell script `aws-secrets-local-setup.sh` to download all the secrets. This will only work if you are logged into AWS and have access the secrets.
3. Create directory called data and include the IP DB file. Ensure the file is called **IPv4-DB11.BIN**.
4. Run the API with `go run .` (serves HTTPS on port `9000`), or build the binary and run it via `docker-compose-local.yaml`.

## Testing

| Command            | Notes        |
| ------------------ | ------------ |
| go test ./...      | Run all tests - no special perks |
| go clean -testcache && go test ./...      | Clear cache and runs all tests again |

There is also a shell script - `test.sh` that can be used to test the API.

## Contact & Support

All info about the project can be found in the [SKC website](https://thesupremekingscastle.com/about)

If you have any suggestions or critiques you can open up a [Git Issue](https://github.com/ygo-skc/skc-suggestion-engine/issues)

This project was made to improve the SKC site and introduce myself to a new programming language. If you want to support, it's real simple (and free) ➡️ subscribe to [my channel on YT](https://www.youtube.com/c/SupremeKing25)!

## Usage

As of now, no one is permitted to use the API in any way (commercial or otherwise). The reason is, I don't have money to support multiple instances and environments. If multiple calls are being made outside of my immediate vision, the performance will degrade and it's original purpose will be compromised.

If you need an API for teaching/education purposes (not commercial), check out the [SKC API](https://github.com/ygo-skc/skc-api#others)

Otherwise, if you want to use this API for your projects, you can expedite the process of multiple instances being spun up by subscribing and watching [my content on YT](https://www.youtube.com/c/SupremeKing25). Again, this is free and will allow me to offset projects like this with YT Monetization.