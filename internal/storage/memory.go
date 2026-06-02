package storage

import (
	"context"
	"errors"
	"sync"

	"dspm/internal/domain"
)

var ErrNotFound = errors.New("not found")

type Repository interface {
	CreateAgent(ctx context.Context, agent domain.Agent) (domain.Agent, error)
	GetAgent(ctx context.Context, id string) (domain.Agent, error)
	ListAgents(ctx context.Context) ([]domain.Agent, error)
	UpdateAgent(ctx context.Context, agent domain.Agent) (domain.Agent, error)
	CreateSkill(ctx context.Context, skill domain.Skill) (domain.Skill, error)
	GetSkill(ctx context.Context, id string) (domain.Skill, error)
	ListSkills(ctx context.Context) ([]domain.Skill, error)
	CreateMCPConfig(ctx context.Context, config domain.MCPConfig) (domain.MCPConfig, error)
	GetMCPConfig(ctx context.Context, id string) (domain.MCPConfig, error)
	ListMCPConfigs(ctx context.Context) ([]domain.MCPConfig, error)
}

type MemoryRepository struct {
	mu         sync.RWMutex
	agents     map[string]domain.Agent
	skills     map[string]domain.Skill
	mcpConfigs map[string]domain.MCPConfig
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		agents:     make(map[string]domain.Agent),
		skills:     make(map[string]domain.Skill),
		mcpConfigs: make(map[string]domain.MCPConfig),
	}
}

func (r *MemoryRepository) CreateAgent(ctx context.Context, agent domain.Agent) (domain.Agent, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.agents[agent.ID] = cloneAgent(agent)
	return cloneAgent(agent), nil
}

func (r *MemoryRepository) GetAgent(ctx context.Context, id string) (domain.Agent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	agent, ok := r.agents[id]
	if !ok {
		return domain.Agent{}, ErrNotFound
	}
	return cloneAgent(agent), nil
}

func (r *MemoryRepository) ListAgents(ctx context.Context) ([]domain.Agent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	agents := make([]domain.Agent, 0, len(r.agents))
	for _, agent := range r.agents {
		agents = append(agents, cloneAgent(agent))
	}
	return agents, nil
}

func (r *MemoryRepository) UpdateAgent(ctx context.Context, agent domain.Agent) (domain.Agent, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.agents[agent.ID]; !ok {
		return domain.Agent{}, ErrNotFound
	}
	r.agents[agent.ID] = cloneAgent(agent)
	return cloneAgent(agent), nil
}

func (r *MemoryRepository) CreateSkill(ctx context.Context, skill domain.Skill) (domain.Skill, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.skills[skill.ID] = skill
	return skill, nil
}

func (r *MemoryRepository) GetSkill(ctx context.Context, id string) (domain.Skill, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	skill, ok := r.skills[id]
	if !ok {
		return domain.Skill{}, ErrNotFound
	}
	return skill, nil
}

func (r *MemoryRepository) ListSkills(ctx context.Context) ([]domain.Skill, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	skills := make([]domain.Skill, 0, len(r.skills))
	for _, skill := range r.skills {
		skills = append(skills, skill)
	}
	return skills, nil
}

func (r *MemoryRepository) CreateMCPConfig(ctx context.Context, config domain.MCPConfig) (domain.MCPConfig, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.mcpConfigs[config.ID] = cloneMCPConfig(config)
	return cloneMCPConfig(config), nil
}

func (r *MemoryRepository) GetMCPConfig(ctx context.Context, id string) (domain.MCPConfig, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	config, ok := r.mcpConfigs[id]
	if !ok {
		return domain.MCPConfig{}, ErrNotFound
	}
	return cloneMCPConfig(config), nil
}

func (r *MemoryRepository) ListMCPConfigs(ctx context.Context) ([]domain.MCPConfig, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	configs := make([]domain.MCPConfig, 0, len(r.mcpConfigs))
	for _, config := range r.mcpConfigs {
		configs = append(configs, cloneMCPConfig(config))
	}
	return configs, nil
}

func cloneAgent(agent domain.Agent) domain.Agent {
	agent.SkillIDs = append([]string(nil), agent.SkillIDs...)
	agent.MCPConfigIDs = append([]string(nil), agent.MCPConfigIDs...)
	return agent
}

func cloneMCPConfig(config domain.MCPConfig) domain.MCPConfig {
	config.Args = append([]string(nil), config.Args...)
	if config.Env != nil {
		env := make(map[string]string, len(config.Env))
		for key, value := range config.Env {
			env[key] = value
		}
		config.Env = env
	}
	return config
}
