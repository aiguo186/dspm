package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"dspm/internal/service"
	"dspm/internal/storage"
)

type Server struct {
	service *service.Service
	mux     *http.ServeMux
}

type response struct {
	Data  any       `json:"data"`
	Error *apiError `json:"error"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func New(service *service.Service) *Server {
	server := &Server{
		service: service,
		mux:     http.NewServeMux(),
	}
	server.routes()
	return server
}

func (s *Server) Handler() http.Handler {
	return http.MaxBytesHandler(s.mux, 1<<20)
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /healthz", s.healthz)
	s.mux.HandleFunc("POST /api/v1/agents", s.createAgent)
	s.mux.HandleFunc("GET /api/v1/agents", s.listAgents)
	s.mux.HandleFunc("GET /api/v1/agents/", s.agentByID)
	s.mux.HandleFunc("POST /api/v1/agents/", s.agentAction)
	s.mux.HandleFunc("POST /api/v1/skills", s.createSkill)
	s.mux.HandleFunc("GET /api/v1/skills", s.listSkills)
	s.mux.HandleFunc("GET /api/v1/skills/", s.skillByID)
	s.mux.HandleFunc("POST /api/v1/mcp-configs", s.createMCPConfig)
	s.mux.HandleFunc("GET /api/v1/mcp-configs", s.listMCPConfigs)
	s.mux.HandleFunc("GET /api/v1/mcp-configs/", s.mcpConfigByID)
}

func (s *Server) healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) createAgent(w http.ResponseWriter, r *http.Request) {
	var req service.CreateAgentRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	agent, err := s.service.CreateAgent(r.Context(), req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, agent)
}

func (s *Server) listAgents(w http.ResponseWriter, r *http.Request) {
	agents, err := s.service.ListAgents(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, agents)
}

func (s *Server) agentByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/agents/")
	if id == "" || strings.Contains(id, "/") {
		writeBadRequest(w, "invalid agent id")
		return
	}
	agent, err := s.service.GetAgent(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, agent)
}

func (s *Server) agentAction(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/agents/")
	parts := strings.Split(path, "/")
	if len(parts) != 3 {
		writeBadRequest(w, "invalid agent action path")
		return
	}
	agentID := parts[0]
	resource := parts[1]
	resourceID := parts[2]
	switch resource {
	case "skills":
		agent, err := s.service.BindSkill(r.Context(), agentID, resourceID)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, agent)
	case "mcp-configs":
		agent, err := s.service.BindMCPConfig(r.Context(), agentID, resourceID)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, agent)
	default:
		writeBadRequest(w, "unsupported agent action")
	}
}

func (s *Server) createSkill(w http.ResponseWriter, r *http.Request) {
	var req service.CreateSkillRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	skill, err := s.service.CreateSkill(r.Context(), req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, skill)
}

func (s *Server) listSkills(w http.ResponseWriter, r *http.Request) {
	skills, err := s.service.ListSkills(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, skills)
}

func (s *Server) skillByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/skills/")
	if id == "" || strings.Contains(id, "/") {
		writeBadRequest(w, "invalid skill id")
		return
	}
	skill, err := s.service.GetSkill(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, skill)
}

func (s *Server) createMCPConfig(w http.ResponseWriter, r *http.Request) {
	var req service.CreateMCPConfigRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	config, err := s.service.CreateMCPConfig(r.Context(), req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, config)
}

func (s *Server) listMCPConfigs(w http.ResponseWriter, r *http.Request) {
	configs, err := s.service.ListMCPConfigs(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, configs)
}

func (s *Server) mcpConfigByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/mcp-configs/")
	if id == "" || strings.Contains(id, "/") {
		writeBadRequest(w, "invalid mcp config id")
		return
	}
	config, err := s.service.GetMCPConfig(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, config)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		writeBadRequest(w, "invalid json body")
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(response{Data: data, Error: nil})
}

func writeError(w http.ResponseWriter, err error) {
	if errors.Is(err, service.ErrInvalidRequest) {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if errors.Is(err, storage.ErrNotFound) {
		writeAPIError(w, http.StatusNotFound, "not_found", "resource not found")
		return
	}
	writeAPIError(w, http.StatusInternalServerError, "internal_error", "internal server error")
}

func writeBadRequest(w http.ResponseWriter, message string) {
	writeAPIError(w, http.StatusBadRequest, "invalid_request", message)
}

func writeAPIError(w http.ResponseWriter, status int, code string, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(response{Data: nil, Error: &apiError{Code: code, Message: message}})
}
