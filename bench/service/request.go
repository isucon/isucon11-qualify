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
	TargetIP   string `json:"target_ip"`
	TargetPort int    `json:"target_port"`
	IsuUUID    string `json:"isu_uuid"`
}

type PostIsuRequest struct {
	JIAIsuUUID string `json:"jia_isu_uuid"`
	IsuName    string `json:"isu_name"`
	Img        []byte
}

type IsuImg struct {
	ImgName string
	Img     []byte
}

type PutIsuRequest struct {
	Name string `json:"name"`
}

type GetIsuConditionRequest struct {
	StartTime      *int64
	EndTime        int64
	ConditionLevel string
	Limit          *int
}

type GetGraphRequest struct {
	Date int64 // unixtime
}
