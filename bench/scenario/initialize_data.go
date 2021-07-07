package scenario

import "github.com/isucon/isucon11-qualify/bench/model"

func (s *Scenario) InitializeData() {
	//TODO: ちゃんと生成する

	s.Catalogs = map[string]*model.IsuCatalog{
		"550e8400-e29b-41d4-a716-446655440000": {
			ID:          "550e8400-e29b-41d4-a716-446655440000",
			Name:        "isu0",
			LimitWeight: 150,
			Weight:      30,
			Size:        "W65.5×D66×H114.5~128.5cm",
			Maker:       "isu maker",
			Features:    "headrest,armrest",
		},
		"562dc0df-2d4f-4e38-98c0-9333f4ff3e38": {
			ID:          "550e8400-e29b-41d4-a716-446655440000",
			Name:        "isu1",
			LimitWeight: 136,
			Weight:      15,
			Size:        "W47×D43×H91cm～97cm",
			Maker:       "isu maker 2",
			Features:    "",
		},
	}
}
