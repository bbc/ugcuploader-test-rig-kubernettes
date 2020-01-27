package types

//RedisTenant used to store infor about tenant in redi
type RedisTenant struct {
	Started string `redis:"started"`
	Errors  string `redis:"errors"`
	Tenant  string `redis:"tenant"`
}

//TestStatus Used to return the status of all running test
type TestStatus struct {
	Started      []RedisTenant
	BeingDeleted []RedisTenant
}

//UgcLoadRequest This is used to map to the form data.. seems to only work with firefox
type UgcLoadRequest struct {
	Context              string `json:"context" form:"context" validate:"required"`
	NumberOfNodes        int    `json:"numberOfNodes" form:"numberOfNodes" validate:"numeric,min=1"`
	BandWidthSelection   string `json:"bandWidthSelection" form:"bandWidthSelection" validate:"required"`
	Jmeter               string `json:"jmeter" form:"jmeter"`
	Data                 string `json:"data" form:"data"`
	MissingTenant        bool
	MissingNumberOfNodes bool
	MissingJmeter        bool
	MissingData          bool
	ProblemsBinding      bool
	MonitorURL           string
	DashboardURL         string
	Success              string
	InvalidTenantName    string
	TenantDeleted        string
	TenantContext        string `json:"TenantContext" form:"TenantContext"`
	TenantMissing        bool
	InvalidTenantDelete  string
	TennantNotDeleted    string
	GenericCreateTestMsg string
	StopContext          string `json:"stopcontext" form:"stopcontext"`
	StopTenantMissing    bool
	InvalidTenantStop    string
	TennantNotStopped    string
	TenantStopped        string
	TenantList           []string
	ReportURL            string
	RunningTests         []Tenant
	AllTenants           []Tenant
}

//Tenant Information about the tenant
type Tenant struct {
	Name      string
	Namespace string
	Running   bool
	PodIP     string
}

//JmeterResponse the response message recieved from the request to the jmeter agent
type JmeterResponse struct {
	Message string
	Code    int
}
