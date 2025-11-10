package models

import "time"

// Favorite represents a saved query
type Favorite struct {
	ID          string    `yaml:"id"`
	Name        string    `yaml:"name"`
	Description string    `yaml:"description"`
	Query       string    `yaml:"query"`
	Tags        []string  `yaml:"tags"`
	Connection  string    `yaml:"connection"`  // Connection name
	Database    string    `yaml:"database"`    // Database name
	CreatedAt   time.Time `yaml:"created_at"`
	UpdatedAt   time.Time `yaml:"updated_at"`
	UsageCount  int       `yaml:"usage_count"`
	LastUsed    time.Time `yaml:"last_used"`
}
