package models

type Product struct {
	OfferId   int    `json:"offer_id"`
	Name      string `json:"name"`
	Price     int    `json:"price"`
	Quantity  int    `json:"quantity"`
	Available bool   `json:"available"`
}

type ProductWithSellerId struct {
	SellerId  int    `json:"seller_id"`
	OfferId   int    `json:"offer_id"`
	Name      string `json:"name"`
	Price     int    `json:"price"`
	Quantity  int    `json:"quantity"`
	Available bool   `json:"available"`
}
