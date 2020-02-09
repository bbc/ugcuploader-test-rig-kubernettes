package types

//RedisTenant used to store infor about tenant in redis
type RedisTenant struct {
	Started string `redis:"started"`
	Errors  string `redis:"errors"`
	Tenant  string `redis:"tenant"`
}

//TestInfo used to hold information about the running tests
type TestInfo struct {
	LocationDatFile      string `redis:"locationDataFile"`
	LocationJmeterScript string `redis:"locationJmeterScript"`
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
	Xms                  string `json:"xms" form:"xms"`
	Xmx                  string `json:"xmx" form:"xmx"`
	CPU                  string `json:"cpu" form:"cpu"`
	RAM                  string `json:"ram" form:"ram"`
	MaxMetaspaceSize     string `json:"maxMetaspaceSize" form:"maxMetaspaceSize"`
	MissingTenant        bool
	MissingNumberOfNodes bool
	MissingJmeter        bool
	MissingData          bool
	ProblemsBinding      bool
	MonitorURL           string
	DashboardURL         string
	ChronografURL        string
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

//SlaveStatus the status of the slave
type SlaveStatus struct {
	Name  string
	Phase string
	PodIP string
}

//NodeCondition The condition of the node
type NodeCondition struct {
	Type              string
	Status            string
	LastHeartbeatTime string
	Reason            string
	Message           string
}

//NodePhase a struct holding the different phases of nodes
type NodePhase struct {
	Phase          string
	Name           string
	InstanceID     string
	NodeConditions []NodeCondition
}

type BandwidthSelection struct {
	HTTPSCPS int
 	HTTPCPS  int
	TIMEOUT  int
}
