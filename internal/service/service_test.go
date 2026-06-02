package service

import (
	"context"
	"testing"

	"dspm/internal/domain"
	"dspm/internal/storage"
)

func TestCreateAndBindAgentDependencies(t *testing.T) {
	svc := New(storage.NewMemoryRepository())
	ctx := context.Background()

	agent, err := svc.CreateAgent(ctx, CreateAgentRequest{Name: "data-assistant", Model: "gpt-4o-mini"})
	if err != nil {
		t.Fatalf("create agent failed: %v", err)
	}

	skill, err := svc.CreateSkill(ctx, CreateSkillRequest{Name: "weather", Type: domain.SkillTypeHTTP, Endpoint: "https://example.com/weather"})
	if err != nil {
		t.Fatalf("create skill failed: %v", err)
	}

	config, err := svc.CreateMCPConfig(ctx, CreateMCPConfigRequest{Name: "filesystem", Transport: domain.MCPTransportStdio, Command: "npx", Args: []string{"-y"}})
	if err != nil {
		t.Fatalf("create mcp config failed: %v", err)
	}

	agent, err = svc.BindSkill(ctx, agent.ID, skill.ID)
	if err != nil {
		t.Fatalf("bind skill failed: %v", err)
	}
	if len(agent.SkillIDs) != 1 || agent.SkillIDs[0] != skill.ID {
		t.Fatalf("unexpected agent skills: %#v", agent.SkillIDs)
	}

	agent, err = svc.BindMCPConfig(ctx, agent.ID, config.ID)
	if err != nil {
		t.Fatalf("bind mcp config failed: %v", err)
	}
	if len(agent.MCPConfigIDs) != 1 || agent.MCPConfigIDs[0] != config.ID {
		t.Fatalf("unexpected agent mcp configs: %#v", agent.MCPConfigIDs)
	}
}

func TestMCPConfigEnvMaskedWhenRead(t *testing.T) {
	svc := New(storage.NewMemoryRepository())
	ctx := context.Background()

	created, err := svc.CreateMCPConfig(ctx, CreateMCPConfigRequest{
		Name:      "remote",
		Transport: domain.MCPTransportSSE,
		URL:       "https://example.com/sse",
		Env:       map[string]string{"API_KEY": "secret"},
	})
	if err != nil {
		t.Fatalf("create mcp config failed: %v", err)
	}

	got, err := svc.GetMCPConfig(ctx, created.ID)
	if err != nil {
		t.Fatalf("get mcp config failed: %v", err)
	}
	if got.Env["API_KEY"] != "******" {
		t.Fatalf("expected masked env, got %#v", got.Env)
	}
}

func TestValidation(t *testing.T) {
	svc := New(storage.NewMemoryRepository())
	ctx := context.Background()

	if _, err := svc.CreateAgent(ctx, CreateAgentRequest{Name: "", Model: "gpt-4o-mini"}); err == nil {
		t.Fatal("expected agent validation error")
	}
	if _, err := svc.CreateSkill(ctx, CreateSkillRequest{Name: "x", Type: "unknown"}); err == nil {
		t.Fatal("expected skill validation error")
	}
	if _, err := svc.CreateMCPConfig(ctx, CreateMCPConfigRequest{Name: "x", Transport: domain.MCPTransportStdio}); err == nil {
		t.Fatal("expected mcp validation error")
	}
}
