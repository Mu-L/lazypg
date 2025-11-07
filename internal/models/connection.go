package models

import (
	"time"
)

// ConnectionConfig represents a PostgreSQL connection configuration
type ConnectionConfig struct {
	Name     string `yaml:"name"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	SSLMode  string `yaml:"ssl_mode"`
}

// Connection represents an active database connection
type Connection struct {
	ID          string
	Config      ConnectionConfig
	Connected   bool
	ConnectedAt time.Time
	LastPing    time.Time
	Error       error
}

// ConnectionState represents the current connection state
type ConnectionState int

const (
	Disconnected ConnectionState = iota
	Connecting
	Connected
	Failed
)

// DiscoveredInstance represents a PostgreSQL instance found via auto-discovery
type DiscoveredInstance struct {
	Host         string
	Port         int
	Source       DiscoverySource
	Available    bool
	ResponseTime time.Duration
}

// DiscoverySource indicates how an instance was discovered
type DiscoverySource int

const (
	SourcePortScan DiscoverySource = iota
	SourceEnvironment
	SourcePgPass
	SourcePgService
	SourceUnixSocket
	SourceConfig
)

func (s DiscoverySource) String() string {
	switch s {
	case SourcePortScan:
		return "Port Scan"
	case SourceEnvironment:
		return "Environment"
	case SourcePgPass:
		return ".pgpass"
	case SourcePgService:
		return ".pg_service.conf"
	case SourceUnixSocket:
		return "Unix Socket"
	case SourceConfig:
		return "Config File"
	default:
		return "Unknown"
	}
}
