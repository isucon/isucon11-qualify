package service

import "github.com/francoispqt/gojay"

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

	Icon           []byte `json:"-"`
	IconStatusCode int    //icon取得時のstatus code(200 or 304想定)
}

type GetIsuConditionResponseArray []GetIsuConditionResponse

type GetIsuConditionResponse struct {
	JIAIsuUUID     string `json:"jia_isu_uuid"`
	IsuName        string `json:"isu_name"`
	Timestamp      int64  `json:"timestamp"`
	IsSitting      bool   `json:"is_sitting"`
	Condition      string `json:"condition"`
	ConditionLevel string `json:"condition_level"`
	Message        string `json:"message"`
}

func (c *GetIsuConditionResponse) UnmarshalJSONObject(dec *gojay.Decoder, key string) error {
	switch key {
	case "jia_isu_uuid":
		return dec.String(&c.JIAIsuUUID)
	case "isu_name":
		return dec.String(&c.IsuName)
	case "timestamp":
		return dec.Int64(&c.Timestamp)
	case "is_sitting":
		return dec.Bool(&c.IsSitting)
	case "condition":
		return dec.String(&c.Condition)
	case "condition_level":
		return dec.String(&c.ConditionLevel)
	case "message":
		return dec.String(&c.Message)
	}
	return nil
}
func (c *GetIsuConditionResponse) NKeys() int {
	return 7
}

func (c *GetIsuConditionResponseArray) UnmarshalJSONArray(dec *gojay.Decoder) error {
	condition := &GetIsuConditionResponse{}
	if err := dec.Object(condition); err != nil {
		return err
	}
	*c = append(*c, *condition)
	return nil
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
	Character string          `json:"character"`
	Info      TrendConditions `json:"info"`
	Warning   TrendConditions `json:"warning"`
	Critical  TrendConditions `json:"critical"`
}

type TrendConditions []TrendCondition

type TrendCondition struct {
	IsuID     int   `json:"isu_id"`
	Timestamp int64 `json:"timestamp"`
}

func (t *TrendCondition) UnmarshalJSONObject(dec *gojay.Decoder, key string) error {
	switch key {
	case "isu_id":
		return dec.Int(&t.IsuID)
	case "timestamp":
		return dec.Int64(&t.Timestamp)
	}
	return nil
}
func (t *TrendCondition) NKeys() int {
	return 2
}

func (t *TrendConditions) UnmarshalJSONArray(dec *gojay.Decoder) error {
	cond := &TrendCondition{}
	if err := dec.Object(cond); err != nil {
		return err
	}
	*t = append(*t, *cond)
	return nil
}

func (t *GetTrendResponseOne) UnmarshalJSONObject(dec *gojay.Decoder, key string) error {
	switch key {
	case "character":
		return dec.String(&t.Character)
	case "info":
		conditions := TrendConditions{}
		if err := dec.Array(&conditions); err != nil {
			return err
		}
		t.Info = conditions
	case "warning":
		conditions := TrendConditions{}
		if err := dec.Array(&conditions); err != nil {
			return err
		}
		t.Warning = conditions
	case "critical":
		conditions := TrendConditions{}
		if err := dec.Array(&conditions); err != nil {
			return err
		}
		t.Critical = conditions
	}
	return nil
}
func (t *GetTrendResponseOne) NKeys() int {
	return 4
}

func (t *GetTrendResponse) UnmarshalJSONArray(dec *gojay.Decoder) error {
	GetTrendResponseOne := &GetTrendResponseOne{}
	if err := dec.Object(GetTrendResponseOne); err != nil {
		return err
	}
	*t = append(*t, *GetTrendResponseOne)
	return nil
}
