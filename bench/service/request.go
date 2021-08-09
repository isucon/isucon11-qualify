package service

type PostInitializeRequest struct {
	JIAServiceURL string `json:"jia_service_url"`
}

type PostIsuConditionRequest struct {
	IsSitting bool   `json:"is_sitting"`
	Condition string `json:"condition"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

type JIAServiceRequest struct {
	TargetBaseURL string `json:"target_base_url"`
	IsuUUID       string `json:"isu_uuid"`
}

type PostIsuRequest struct {
	JIAIsuUUID string `json:"jia_isu_uuid"`
	IsuName    string `json:"isu_name"`
	Img        []byte
}

type GetIsuConditionRequest struct {
	StartTime      *int64
	EndTime        int64
	ConditionLevel string
}

type GetGraphRequest struct {
	Date int64 // unixtime
}
