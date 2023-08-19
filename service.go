package fuse

// Service is the interface that all services must implement
// It defines the methods that are called when the service is started or stopped
type Service interface {
	// Create creates a new instance of the service for a provided guild manager
	Create(mng *GuildManager) (Service, error)
	// Start is called when the service is started
	Start(mng *GuildManager) error
	// Stop is called when the service is stopped
	Stop(mng *GuildManager) error
}

// ServiceConfiguration is a simple struct used to store a service's configuration in the database
// This is not always required but has a configured guild ID field for convenience
type ServiceConfiguration struct {
	// GuildID is the ID of the guild and is used as the primary key
	GuildId string `gorm:"primary_key"`
}
