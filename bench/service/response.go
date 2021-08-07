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
	ID                 int                      `json:"id"`
	JIAIsuUUID         string                   `json:"jia_isu_uuid"`
	Name               string                   `json:"name"`
	Character          string                   `json:"character"`
	LatestIsuCondition *GetIsuConditionResponse `json:"latest_isu_condition"`
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
	Score      int                 `json:"score"`
	Percentage GraphDataPercentage `json:"percentage"`
}

type GraphDataPercentage struct {
	Sitting      int `json:"sitting"`
	IsBroken     int `json:"is_broken"`
	IsDirty      int `json:"is_dirty"`
	IsOverweight int `json:"is_overweight"`
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
