package contracts

type CardInfoResponse struct {
	CardID        string `json:"cardID"`
	CardName      string `json:"cardName"`
	CardColor     string `json:"cardColor"`
	CardAttribute string `json:"cardAttribute"`
	CardEffect    string `json:"cardEffect"`
}

const (
	SkcBaseUrl       = "https://skc-ygo-api.com"
	CardInfoEndpoint = "/api/v1/card/67288539?allInfo=true"
)
