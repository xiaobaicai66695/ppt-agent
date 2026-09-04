package metrics

// RecordTaskCreated 增加创建任务计数器（指定状态）
func RecordTaskCreated() {
	TasksTotal.WithLabelValues("running").Inc()
	ActiveTasks.Inc()
}

// RecordTaskCompleted 任务完成时记录指标
func RecordTaskCompleted(durationSeconds float64, slidesGenerated, slidesTotal int, status string) {
	TasksTotal.WithLabelValues(status).Inc()
	ActiveTasks.Dec()
	TaskDuration.WithLabelValues(status).Observe(durationSeconds)
	if slidesGenerated > 0 {
		TaskSlidesGenerated.WithLabelValues(status).Observe(float64(slidesGenerated))
	}
}

// RecordTokens 记录 token 使用量
func RecordTokens(prompt, completion, total int64) {
	LLMTokensTotal.WithLabelValues("prompt").Add(float64(prompt))
	LLMTokensTotal.WithLabelValues("completion").Add(float64(completion))
	LLMTokensTotal.WithLabelValues("total").Add(float64(total))
}

// RecordLLMCall 记录一次 LLM 调用结果
func RecordLLMCall(status string) {
	LLMCallsTotal.WithLabelValues(status).Inc()
}

// RecordAgentCall 记录一次代理调用
func RecordAgentCall(agent string) {
	AgentCallsTotal.WithLabelValues(agent).Inc()
}

// RecordToolCall 记录一次工具调用
func RecordToolCall(tool, status string) {
	ToolCallsTotal.WithLabelValues(tool, status).Inc()
}

// RecordQAIssue 记录一个发现的 QA 问题
func RecordQAIssue(severity string) {
	QAIssuesTotal.WithLabelValues(severity).Inc()
}

// RecordQAFix 记录一次 QA 修复尝试结果
func RecordQAFix(success bool) {
	if success {
		QAFixesTotal.WithLabelValues("success").Inc()
	} else {
		QAFixesTotal.WithLabelValues("failed").Inc()
	}
}

// RecordSlideScore 记录 QA 审查后的每张幻灯片质量评分
func RecordSlideScore(score float64, contentType string) {
	QASlideScore.WithLabelValues(contentType).Observe(score)
}
