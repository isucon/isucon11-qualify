package service

type InitializeResponse struct {
	Language string `json:"language"`
}

type AuthResponse struct {
}
type SignoutResponse struct {
}

type GetMeResponse struct {
	JIAUserID string `json:"jia_user_id"`
}

type Isu struct {
	JIAIsuUUID string `json:"jia_isu_uuid"`
	Name       string `json:"name"`
	Character  string `json:"character"`
	// TODO: これはmodelの方にあるのが正しそう
	Icon []byte `json:"-"`
}

type Catalog struct {
	JIACatalogID string `json:"jia_catalog_id"`
	Name         string `json:"name"`
	LimitWeight  int    `json:"limit_weight"`
	Weight       int    `json:"weight"`
	Size         string `json:"size"`
	Maker        string `json:"maker"`
	Tags         string `json:"tags"`
}

type GetIsuConditionResponse struct {
	JIAIsuUUID     string `json:"jia_isu_uuid"`
	IsuName        string `json:"isu_name"`
	Timestamp      int64  `json:"timestamp"`
	IsSitting      bool   `json:"is_sitting"`
	Condition      string `json:"condition"`
	ConditionLevel string `json:"condition_level"`
	Message        string `json:"message"`
}

type GraphResponse struct {
	StartAt int64      `json:"start_at"`
	EndAt   int64      `json:"end_at"`
	Data    *GraphData `json:"data"`
}

type GraphData struct {
	Score   int            `json:"score"`
	Sitting int            `json:"sitting"`
	Detail  map[string]int `json:"detail"`
}
