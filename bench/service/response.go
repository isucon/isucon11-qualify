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
	ID         int    `json:"id"`
	JIAIsuUUID string `json:"jia_isu_uuid"`
	Name       string `json:"name"`
	Character  string `json:"character"`
	// TODO: これはmodelの方にあるのが正しそう
	Icon []byte `json:"-"`
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

type GraphResponse []*GraphResponseOne

type GraphResponseOne struct {
	StartAt             int64      `json:"start_at"`
	EndAt               int64      `json:"end_at"`
	Data                *GraphData `json:"data"`
	ConditionTimestamps []int64    `json:"condition_timestamps"`
}

type GraphData struct {
	Score   int            `json:"score"`
	Sitting int            `json:"sitting"`
	Detail  map[string]int `json:"detail"`
}

type GetTrendResponse []GetTrendResponseOne

type GetTrendResponseOne struct {
	Character  string           `json:"character"`
	Conditions []TrendCondition `json:"conditions"`
}

type TrendCondition struct {
	IsuID          int    `json:"isu_id"`
	Timestamp      int64  `json:"timestamp"`
	ConditionLevel string `json:"condition_level"`
}
