package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"dspm/internal/domain"
	"dspm/internal/storage"
)

var ErrInvalidRequest = errors.New("invalid request")

type Service struct {
	repo storage.Repository
}

type CreateAgentRequest struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Model        string `json:"model"`
	SystemPrompt string `json:"system_prompt"`
}

type CreateSkillRequest struct {
	Name         string           `json:"name"`
	Description  string           `json:"description"`
	Type         domain.SkillType `json:"type"`
	Endpoint     string           `json:"endpoint"`
	InputSchema  string           `json:"input_schema"`
	OutputSchema string           `json:"output_schema"`
}

type CreateMCPConfigRequest struct {
	Name      string              `json:"name"`
	Transport domain.MCPTransport `json:"transport"`
	Command   string              `json:"command"`
	Args      []string            `json:"args"`
	URL       string              `json:"url"`
	Env       map[string]string   `json:"env"`
}

func New(repo storage.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateAgent(ctx context.Context, req CreateAgentRequest) (domain.Agent, error) {
	name := strings.TrimSpace(req.Name)
	model := strings.TrimSpace(req.Model)
	if name == "" {
		return domain.Agent{}, fmt.Errorf("%w: name is required", ErrInvalidRequest)
	}
	if model == "" {
		return domain.Agent{}, fmt.Errorf("%w: model is required", ErrInvalidRequest)
	}
	now := time.Now().UTC()
	agent := domain.Agent{
		ID:           newID("agt"),
		Name:         name,
		Description:  strings.TrimSpace(req.Description),
		Model:        model,
		SystemPrompt: strings.TrimSpace(req.SystemPrompt),
		SkillIDs:     []string{},
		MCPConfigIDs: []string{},
		Status:       domain.StatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	return s.repo.CreateAgent(ctx, agent)
}

func (s *Service) GetAgent(ctx context.Context, id string) (domain.Agent, error) {
	return s.repo.GetAgent(ctx, id)
}

func (s *Service) ListAgents(ctx context.Context) ([]domain.Agent, error) {
	return s.repo.ListAgents(ctx)
}

func (s *Service) BindSkill(ctx context.Context, agentID string, skillID string) (domain.Agent, error) {
	agent, err := s.repo.GetAgent(ctx, agentID)
	if err != nil {
		return domain.Agent{}, err
	}
	skill, err := s.repo.GetSkill(ctx, skillID)
	if err != nil {
		return domain.Agent{}, err
	}
	if skill.Status != domain.StatusActive {
		return domain.Agent{}, fmt.Errorf("%w: skill is disabled", ErrInvalidRequest)
	}
	if !contains(agent.SkillIDs, skillID) {
		agent.SkillIDs = append(agent.SkillIDs, skillID)
		agent.UpdatedAt = time.Now().UTC()
	}
	return s.repo.UpdateAgent(ctx, agent)
}

func (s *Service) BindMCPConfig(ctx context.Context, agentID string, mcpConfigID string) (domain.Agent, error) {
	agent, err := s.repo.GetAgent(ctx, agentID)
	if err != nil {
		return domain.Agent{}, err
	}
	config, err := s.repo.GetMCPConfig(ctx, mcpConfigID)
	if err != nil {
		return domain.Agent{}, err
	}
	if config.Status != domain.StatusActive {
		return domain.Agent{}, fmt.Errorf("%w: mcp config is disabled", ErrInvalidRequest)
	}
	if !contains(agent.MCPConfigIDs, mcpConfigID) {
		agent.MCPConfigIDs = append(agent.MCPConfigIDs, mcpConfigID)
		agent.UpdatedAt = time.Now().UTC()
	}
	return s.repo.UpdateAgent(ctx, agent)
}

func (s *Service) CreateSkill(ctx context.Context, req CreateSkillRequest) (domain.Skill, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return domain.Skill{}, fmt.Errorf("%w: name is required", ErrInvalidRequest)
	}
	if !validSkillType(req.Type) {
		return domain.Skill{}, fmt.Errorf("%w: unsupported skill type", ErrInvalidRequest)
	}
	now := time.Now().UTC()
	skill := domain.Skill{
		ID:           newID("skl"),
		Name:         name,
		Description:  strings.TrimSpace(req.Description),
		Type:         req.Type,
		Endpoint:     strings.TrimSpace(req.Endpoint),
		InputSchema:  strings.TrimSpace(req.InputSchema),
		OutputSchema: strings.TrimSpace(req.OutputSchema),
		Status:       domain.StatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	return s.repo.CreateSkill(ctx, skill)
}

func (s *Service) GetSkill(ctx context.Context, id string) (domain.Skill, error) {
	return s.repo.GetSkill(ctx, id)
}

func (s *Service) ListSkills(ctx context.Context) ([]domain.Skill, error) {
	return s.repo.ListSkills(ctx)
}

func (s *Service) CreateMCPConfig(ctx context.Context, req CreateMCPConfigRequest) (domain.MCPConfig, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return domain.MCPConfig{}, fmt.Errorf("%w: name is required", ErrInvalidRequest)
	}
	if !validMCPTransport(req.Transport) {
		return domain.MCPConfig{}, fmt.Errorf("%w: unsupported mcp transport", ErrInvalidRequest)
	}
	if req.Transport == domain.MCPTransportStdio && strings.TrimSpace(req.Command) == "" {
		return domain.MCPConfig{}, fmt.Errorf("%w: command is required for stdio transport", ErrInvalidRequest)
	}
	if req.Transport != domain.MCPTransportStdio && strings.TrimSpace(req.URL) == "" {
		return domain.MCPConfig{}, fmt.Errorf("%w: url is required for remote transport", ErrInvalidRequest)
	}
	now := time.Now().UTC()
	config := domain.MCPConfig{
		ID:        newID("mcp"),
		Name:      name,
		Transport: req.Transport,
		Command:   strings.TrimSpace(req.Command),
		Args:      append([]string(nil), req.Args...),
		URL:       strings.TrimSpace(req.URL),
		Env:       cloneEnv(req.Env),
		Status:    domain.StatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
	return s.repo.CreateMCPConfig(ctx, config)
}

func (s *Service) GetMCPConfig(ctx context.Context, id string) (domain.MCPConfig, error) {
	config, err := s.repo.GetMCPConfig(ctx, id)
	if err != nil {
		return domain.MCPConfig{}, err
	}
	config.Env = maskEnv(config.Env)
	return config, nil
}

func (s *Service) ListMCPConfigs(ctx context.Context) ([]domain.MCPConfig, error) {
	configs, err := s.repo.ListMCPConfigs(ctx)
	if err != nil {
		return nil, err
	}
	for i := range configs {
		configs[i].Env = maskEnv(configs[i].Env)
	}
	return configs, nil
}

func validSkillType(value domain.SkillType) bool {
	switch value {
	case domain.SkillTypeBuiltin, domain.SkillTypeHTTP, domain.SkillTypeMCPTool, domain.SkillTypeWorkflow:
		return true
	default:
		return false
	}
}

func validMCPTransport(value domain.MCPTransport) bool {
	switch value {
	case domain.MCPTransportStdio, domain.MCPTransportSSE, domain.MCPTransportStreamableHTTP:
		return true
	default:
		return false
	}
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func cloneEnv(env map[string]string) map[string]string {
	if env == nil {
		return map[string]string{}
	}
	cloned := make(map[string]string, len(env))
	for key, value := range env {
		cloned[key] = value
	}
	return cloned
}

func maskEnv(env map[string]string) map[string]string {
	masked := make(map[string]string, len(env))
	for key := range env {
		masked[key] = "******"
	}
	return masked
}

func newID(prefix string) string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
	}
	return prefix + "_" + hex.EncodeToString(buf)
}
