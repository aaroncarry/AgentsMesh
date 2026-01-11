package runner

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// HostInfo represents runner host information
type HostInfo map[string]interface{}

// Scan implements sql.Scanner for HostInfo
func (hi *HostInfo) Scan(value interface{}) error {
	if value == nil {
		*hi = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, hi)
}

// Value implements driver.Valuer for HostInfo
func (hi HostInfo) Value() (driver.Value, error) {
	if hi == nil {
		return nil, nil
	}
	return json.Marshal(hi)
}

// RegistrationToken represents a token used to register runners
type RegistrationToken struct {
	ID             int64  `gorm:"primaryKey" json:"id"`
	OrganizationID int64  `gorm:"not null;index" json:"organization_id"`
	TokenHash      string `gorm:"size:255;not null;uniqueIndex" json:"-"`
	Description    string `gorm:"type:text" json:"description,omitempty"`
	CreatedByID    int64  `gorm:"not null" json:"created_by_id"`

	IsActive  bool       `gorm:"not null;default:true" json:"is_active"`
	MaxUses   *int       `json:"max_uses,omitempty"`
	UsedCount int        `gorm:"not null;default:0" json:"used_count"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`

	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`
}

func (RegistrationToken) TableName() string {
	return "runner_registration_tokens"
}

// Runner status constants
const (
	RunnerStatusOnline  = "online"
	RunnerStatusOffline = "offline"
	RunnerStatusBusy    = "busy"
)

// Runner represents a self-hosted runner
type Runner struct {
	ID             int64  `gorm:"primaryKey" json:"id"`
	OrganizationID int64  `gorm:"not null;index" json:"organization_id"`
	NodeID         string `gorm:"size:100;not null" json:"node_id"`
	Description    string `gorm:"type:text" json:"description,omitempty"`
	AuthTokenHash  string `gorm:"size:255;not null" json:"-"`

	Status            string     `gorm:"size:50;not null;default:'offline';index" json:"status"`
	LastHeartbeat     *time.Time `json:"last_heartbeat,omitempty"`
	CurrentPods       int        `gorm:"not null;default:0" json:"current_pods"`
	MaxConcurrentPods int        `gorm:"not null;default:5" json:"max_concurrent_pods"`
	RunnerVersion     *string    `gorm:"size:50" json:"runner_version,omitempty"`
	IsEnabled             bool       `gorm:"not null;default:true" json:"is_enabled"`

	HostInfo     HostInfo     `gorm:"type:jsonb" json:"host_info,omitempty"`
	Capabilities Capabilities `gorm:"type:jsonb" json:"capabilities,omitempty"`

	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null;default:now()" json:"updated_at"`
}

// Capabilities represents runner plugin capabilities (JSONB type)
type Capabilities []PluginCapability

// Scan implements sql.Scanner for Capabilities
func (c *Capabilities) Scan(value interface{}) error {
	if value == nil {
		*c = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, c)
}

// Value implements driver.Valuer for Capabilities
func (c Capabilities) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	return json.Marshal(c)
}

// PluginCapability represents a single plugin's capability
type PluginCapability struct {
	Name            string    `json:"name"`
	Version         string    `json:"version"`
	Description     string    `json:"description"`
	SupportedAgents []string  `json:"supported_agents"`
	Executable      string    `json:"executable,omitempty"` // Required CLI command (if any)
	Available       bool      `json:"available"`            // Whether the executable is available on this system
	UI              *UIConfig `json:"ui,omitempty"`
}

// UIConfig represents the UI configuration for a plugin
type UIConfig struct {
	Configurable bool      `json:"configurable"`
	Fields       []UIField `json:"fields"`
}

// UIField represents a single UI field configuration
type UIField struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"` // boolean, string, select, number, secret
	Label       string      `json:"label"`
	Default     interface{} `json:"default,omitempty"`
	Description string      `json:"description,omitempty"`
	Placeholder string      `json:"placeholder,omitempty"`
	Options     []UIOption  `json:"options,omitempty"`
	Min         *float64    `json:"min,omitempty"`
	Max         *float64    `json:"max,omitempty"`
	Required    bool        `json:"required,omitempty"`
}

// UIOption represents an option for select fields
type UIOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// UIFieldType defines valid field types for plugin UI
type UIFieldType string

const (
	UIFieldTypeBoolean UIFieldType = "boolean"
	UIFieldTypeString  UIFieldType = "string"
	UIFieldTypeSelect  UIFieldType = "select"
	UIFieldTypeNumber  UIFieldType = "number"
	UIFieldTypeSecret  UIFieldType = "secret"
)

// Validate validates a PluginCapability
func (p *PluginCapability) Validate() error {
	if p.Name == "" {
		return errors.New("plugin capability name is required")
	}
	return nil
}

// ValidateCapabilities validates a list of capabilities
func ValidateCapabilities(caps []PluginCapability) error {
	for i, cap := range caps {
		if err := cap.Validate(); err != nil {
			return fmt.Errorf("capability[%d]: %w", i, err)
		}
	}
	return nil
}

func (Runner) TableName() string {
	return "runners"
}

// IsOnline returns true if runner is online
func (r *Runner) IsOnline() bool {
	return r.Status == RunnerStatusOnline
}

// CanAcceptPod returns true if runner can accept new pods
func (r *Runner) CanAcceptPod() bool {
	return r.IsEnabled && r.IsOnline() && r.CurrentPods < r.MaxConcurrentPods
}
