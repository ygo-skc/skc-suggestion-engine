package model

type Product struct {
	ProductID          string           `json:"productId"`
	ProductLocale      string           `json:"productLocale"`
	ProductName        string           `json:"productName"`
	ProductType        string           `json:"productType"`
	ProductSubType     string           `json:"productSubType"`
	ProductReleaseDate string           `json:"productReleaseDate"`
	ProductTotal       int              `json:"productTotal,omitempty"`
	ProductRarityStats map[string]int   `json:"productRarityStats,omitempty"`
	ProductContent     []ProductContent `json:"productContent,omitempty"`
}

type ProductContent struct {
	Card            Card     `json:"card"`
	ProductPosition string   `json:"productPosition"`
	Rarities        []string `json:"rarities"`
}
