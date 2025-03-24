package util

type ServiceType int

const (
	BackendService ServiceType = iota
	WebService
	ConnectionsService
	ETLService
	SimulateService
)

func (s ServiceType) GetName() string {
	switch s {
	case BackendService:
		return "BackendService"
	case WebService:
		return "WebService"
	case ConnectionsService:
		return "ConnectionsService"
	case ETLService:
		return "ETLService"
	case SimulateService:
		return "SimulateService"
	default:
		return "Unknown"
	}
}
