package icp

type Options struct {
	ProxyURL string
	Server   string
	Retries  int
}

// Response 表示通用响应结构体
type Response struct {
	Status  string   `json:"status"`
	Data    []*Entry `json:"data,omitempty"`    // 成功时包含的数据，omitempty 表示当值为空时忽略
	Message string   `json:"message,omitempty"` // 错误时包含的详细信息
}

// Entry 表示成功响应中的具体数据项
type Entry struct {
	ContentTypeName  string `json:"contentTypeName"`
	Domain           string `json:"domain"`
	DomainID         int64  `json:"domainId"`
	LeaderName       string `json:"leaderName"`
	LimitAccess      string `json:"limitAccess"`
	MainID           int64  `json:"mainId"`
	MainLicence      string `json:"mainLicence"`
	NatureName       string `json:"natureName"`
	ServiceID        int64  `json:"serviceId"`
	ServiceLicence   string `json:"serviceLicence"`
	UnitName         string `json:"unitName"`
	UpdateRecordTime string `json:"updateRecordTime"` // 使用 time.Time 解析日期
}
