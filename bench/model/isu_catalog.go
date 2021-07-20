package model

type IsuCatalog struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	LimitWeight int64  `json:"limit_weight"`
	Weight      int64  `json:"weight"`
	Size        string `json:"size"`
	Maker       string `json:"maker"`
	Features    string `json:"features"`
}
