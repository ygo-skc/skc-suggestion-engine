# skc-suggestion-engine

## Info

Go API that will extend functionality of [SKC API](https://github.com/ygo-skc/skc-api) with the following:

* **Complete** Allow storage of deck list so my subscribers on YT can copy or view a deck I'm showcasing on a video.
  * Functionality might be extended in the future to allow other people to submit their decks. It all depends on how much time I have to implement this and how many users want this to be added.
* **Complete** Allow clients to send info about cards or products they are browsing with geo-location (using IP address) to build a trending view. This trending view will allow users to see what others are browsing (which helps w/ suggestions) and also see what is trending near them in close to real time.
* **In Progress (Testing Needed)** Allow users to see what materials a specific extra deck monster can use in order to fulfill its summoning conditions. First release will only support direct references (not archetypal)
  * Functionality will be extended to include reference suggestions, ie; if a card mentions another specific card by name (or archetype) the suggestion-engine will return info on all cards mentioned. This will group cards with a relation of sorts. This relation will help users browse similar cards. This will be great as it can leverage further work on cards that search for specific cards.
  * Another possible addition to functionality, suggest cards found together in deck lists. For example: user want suggestions for Card **XXX** and looking at deck lists we see **XXX** and **YYY** are used together more often then not, then we will suggest card **YYY**

## Local Setup
In order for the API to work locally, do the following steps

1. Run `go mod tidy` to download deps
2. Execute the shell **script doppler-secrets-local-setup.sh** to download all the secrets. This will only work if you are logged into Doppler and have access to my org.
3. Create directory called data and include the IP DB file. Ensure the file is called **IPv4-DB.BIN**.

## Contact & Support

All info about the project can be found in the [SKC website](https://thesupremekingscastle.com/about)

If you have any suggestions or critiques you can open up a [Git Issue](https://github.com/ygo-skc/skc-suggestion-engine/issues)

This project was made to improve the SKC site and introduce myself to a new programming language. If you want to support, it's real simple (and free) ➡️ subscribe to [my channel on YT](https://www.youtube.com/c/SupremeKing25)!

## Usage

As of now, no one is permitted to use the API in any way (commercial or otherwise). The reason is, I don't have money to support multiple instances and environments. If multiple calls are being made outside of my immediate vision, the performance will degrade and it's original purpose will be compromised.

If you need an API for teaching/education purposes (not commercial), check out the [SKC API](https://github.com/ygo-skc/skc-api#others)

Otherwise, if you want to use this API for your projects, you can expedite the process of multiple instances being spun up by subscribing and watching [my content on YT](https://www.youtube.com/c/SupremeKing25). Again, this is free and will allow me to offset projects like this with YT Monetization.