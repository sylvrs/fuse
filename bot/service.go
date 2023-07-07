package bot

// Service is the interface that all services must implement
// It defines the methods that are called when the service is started or stopped
type Service interface {
	// Name returns the name of the service
	Name() string
	// Create creates a new instance of the service for a provided guild manager
	Create(mng *GuildManager) (Service, error)
	// Start is called when the service is started
	Start(mng *GuildManager) error
	// Stop is called when the service is stopped
	Stop(mng *GuildManager) error
}
