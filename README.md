# skc-suggestion-engine

[![Unit Test](https://github.com/ygo-skc/skc-suggestion-engine/actions/workflows/unit-test.yaml/badge.svg)](https://github.com/ygo-skc/skc-suggestion-engine/actions/workflows/unit-test.yaml)

## Info

Go API that extends functionality of [SKC API](https://github.com/ygo-skc/skc-api) with the following:

* Allow storage of deck lists - currently not everyone can submit a deck list, functionality might be opened later.
* Clients can send browsing data to build a suggestion database. Clients need to be authenticated to do this.
* Suggest materials by parsing text of a card
* Suggest references by parsing text of a card
* Suggest support cards by analyzing every card in the DB

## Local Setup

In order for the API to work locally, do the following steps

1. Run `go mod tidy` to download deps
2. Execute the shell **script doppler-secrets-local-setup.sh** to download all the secrets. This will only work if you are logged into Doppler and have access to my org.
3. Create directory called data and include the IP DB file. Ensure the file is called **IPv4-DB.BIN**.

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