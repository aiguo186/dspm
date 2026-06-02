package domain

import "time"

type Status string

const (
	StatusActive   Status = "active"
	StatusDisabled Status = "disabled"
)

type SkillType string

const (
	SkillTypeBuiltin  SkillType = "builtin"
	SkillTypeHTTP     SkillType = "http"
	SkillTypeMCPTool  SkillType = "mcp_tool"
	SkillTypeWorkflow SkillType = "workflow"
)

type MCPTransport string

const (
	MCPTransportStdio          MCPTransport = "stdio"
	MCPTransportSSE            MCPTransport = "sse"
	MCPTransportStreamableHTTP MCPTransport = "streamable_http"
)

type Agent struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Model        string    `json:"model"`
	SystemPrompt string    `json:"system_prompt"`
	SkillIDs     []string  `json:"skill_ids"`
	MCPConfigIDs []string  `json:"mcp_config_ids"`
	Status       Status    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Skill struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Type         SkillType `json:"type"`
	Endpoint     string    `json:"endpoint"`
	InputSchema  string    `json:"input_schema"`
	OutputSchema string    `json:"output_schema"`
	Status       Status    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type MCPConfig struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Transport MCPTransport      `json:"transport"`
	Command   string            `json:"command"`
	Args      []string          `json:"args"`
	URL       string            `json:"url"`
	Env       map[string]string `json:"env"`
	Status    Status            `json:"status"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}
