package models

type Storefront struct {
	Name             string `json:"name"`
	URL              string `json:"url"`
	Paks             []Pak  `json:"paks"`
	ExperimentalPaks []Pak  `json:"experimental_paks,omitempty"`
}
