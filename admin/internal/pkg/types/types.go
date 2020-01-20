package types

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
}
