package server

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func (s *Server) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Query string `json:"query"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Query) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "query is required"})
		return
	}

	taskID := s.taskIDGen()
	cfg := s.makeTaskConfig(taskID)
	info, err := s.tasks.CreateTask(r.Context(), req.Query, s.agentFactory, cfg)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, info)
}

func (s *Server) handleGetTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	info := s.tasks.GetTask(id)
	if info == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "task not found"})
		return
	}

	// Enrich with current tasks.json data if running.
	if info.Status == TaskStatusRunning {
		manifest, err := s.tasks.ReadTasksManifestFile(id)
		if err == nil && manifest != nil {
			info.DoneCount = manifest.CompletedCount()
			info.TotalCount = len(manifest.Tasks)
		}
	}

	writeJSON(w, http.StatusOK, info)
}

func (s *Server) handleListTasks(w http.ResponseWriter, r *http.Request) {
	tasks := s.tasks.ListTasks()
	if tasks == nil {
		tasks = []TaskInfo{}
	}
	writeJSON(w, http.StatusOK, tasks)
}

func (s *Server) handleDownloadFile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	filename := r.PathValue("filename")

	ts := s.tasks.GetTaskState(id)
	if ts == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "task not found"})
		return
	}

	filePath := filepath.Join(ts.Info.WorkDir, filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "file not found"})
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.presentationml.presentation")
	http.ServeFile(w, r, filePath)
}

func (s *Server) handleCancelTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if s.tasks.CancelTask(id) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
		return
	}
	writeJSON(w, http.StatusNotFound, map[string]string{"error": "task not found or not running"})
}

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}
