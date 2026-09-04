package web

import webservice "github.com/cloudwego/ppt-agent/pkg/runtime/web/service"

// taskApplication keeps the compatibility Server façade small while giving
// handlers a single application-service boundary. Tests that construct Server
// literals still receive a lazily wired service with the same dependencies.
func (s *Server) taskApplication() *webservice.TaskService {
	if s.taskService != nil {
		return s.taskService
	}
	return webservice.NewTaskService(webservice.TaskServiceConfig{
		Tasks:          s.tasks,
		Sessions:       s.sessionManager,
		AgentFactory:   s.agentFactory,
		MakeTaskConfig: s.makeTaskConfig,
		TemplateLoader: s.templateLoader,
	})
}
