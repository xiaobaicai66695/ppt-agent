/*
 * Copyright 2025 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package learning

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/ppt-agent/pkg/agent/intent"
	"github.com/cloudwego/ppt-agent/pkg/agent/router"
	"github.com/cloudwego/ppt-agent/pkg/logger"
	"github.com/cloudwego/ppt-agent/pkg/style"
)

// Engine 智能学习引擎，整合意图识别、路由、偏好学习和模式分析
type Engine struct {
	classifier  *intent.Classifier
	router      *router.Engine
	collector   *Collector
	updater     *Updater
	analyzer    *Analyzer
	profileStore *style.EnhancedProfileStore

	mu sync.RWMutex
}

// EngineConfig 引擎配置
type EngineConfig struct {
	EnableLLMClassification bool   // 是否启用LLM意图分类
	EnableLearning          bool   // 是否启用持续学习
	EnableProfileMatch      bool   // 是否启用画像匹配
}

// NewEngine 创建智能学习引擎
// modelFactory 用于创建 LLM 实例（用于意图分类等辅助任务）
// 支持两种签名：func(ctx context.Context) (model.ChatModel, error) 或
// func(ctx context.Context) (interface{ Generate(...) }, error)
// 如果传入 nil，则意图分类器只使用规则匹配
func NewEngine(cfg *EngineConfig, modelFactory interface{}) *Engine {
	if cfg == nil {
		cfg = &EngineConfig{
			EnableLLMClassification: true,
			EnableLearning:         true,
			EnableProfileMatch:     true,
		}
	}

	e := &Engine{}

	// 初始化增强画像存储（先于其他组件）
	e.profileStore = style.NewEnhancedProfileStore()

	// 初始化意图分类器（可选择是否启用 LLM）
	e.classifier = intent.NewClassifier(e.makeClassifierFactory(modelFactory))

	// 初始化路由引擎
	e.router = router.NewEngine(e.classifier)

	// 初始化学习组件
	if cfg.EnableLearning {
		e.updater = NewUpdater(e.profileStore)
		e.collector = NewCollector(nil)
		e.analyzer = NewAnalyzer()
		// 建立 Collector → Updater 的连接
		e.collector.SetUpdater(e.updater)
	}

	logger.Info("intelligent_engine_initialized",
		"llm_classification", cfg.EnableLLMClassification,
		"learning", cfg.EnableLearning,
		"profile_match", cfg.EnableProfileMatch,
		"llm_enabled", modelFactory != nil)

	return e
}

// makeClassifierFactory 将通用 modelFactory 适配为 intent.Classifier 需要的类型
func (e *Engine) makeClassifierFactory(modelFactory interface{}) func(ctx context.Context) (model.ToolCallingChatModel, error) {
	if modelFactory == nil {
		return nil
	}

	switch f := modelFactory.(type) {
	case func(ctx context.Context) (model.ToolCallingChatModel, error):
		return f
	case func(ctx context.Context) (model.ChatModel, error):
		// ChatModel 已废弃且不实现 ToolCallingChatModel，不可直接转型
		return nil
	default:
		return func(ctx context.Context) (model.ToolCallingChatModel, error) {
			result, err := callModelFactory(f, ctx)
			if err != nil {
				return nil, err
			}
			if tcm, ok := result.(model.ToolCallingChatModel); ok {
				return tcm, nil
			}
			if cm, ok := result.(model.ChatModel); ok {
				// 尝试通过嵌入 ChatModel 实现 ToolCallingChatModel
				if tcm, ok := any(cm).(model.ToolCallingChatModel); ok {
					return tcm, nil
				}
			}
			return nil, fmt.Errorf("modelFactory 返回值无法转换为 model.ToolCallingChatModel: %T", result)
		}
	}
}

// callModelFactory 动态调用 modelFactory（支持多种函数签名）
func callModelFactory(modelFactory interface{}, ctx context.Context) (interface{}, error) {
	fn := reflect.ValueOf(modelFactory)
	fnType := fn.Type()
	if fnType.Kind() != reflect.Func || fnType.NumIn() != 1 || fnType.NumOut() != 2 {
		return nil, fmt.Errorf("modelFactory 必须是单参数双返回的函数")
	}
	ctxType := fnType.In(0)
	if !ctxType.Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) {
		return nil, fmt.Errorf("modelFactory 参数必须实现 context.Context")
	}

	ctxVal := reflect.ValueOf(ctx)
	result := fn.Call([]reflect.Value{ctxVal})
	if !result[1].IsNil() {
		return nil, result[1].Interface().(error)
	}
	return result[0].Interface(), nil
}

// ProcessTask 处理任务入口
func (e *Engine) ProcessTask(ctx context.Context, query string, userID int) (*ProcessingResult, error) {
	result := &ProcessingResult{
		UserID:    userID,
		Query:     query,
		Timestamp: now(),
	}

	// Step 1: 意图识别
	intentResult, err := e.classifier.Classify(ctx, query, userID)
	if err != nil {
		logger.Warn("intent_classification_failed", "error", err.Error())
		result.Error = err.Error()
	} else {
		result.Intent = intentResult
	}

	// Step 2: 获取用户画像
	profile := e.profileStore.GetEnhanced(userID)
	result.Profile = profile

	// Step 3: 路由决策
	routingResult, err := e.router.Route(ctx, query, userID, profile)
	if err != nil {
		logger.Warn("routing_failed", "error", err.Error())
		result.Error = err.Error()
	} else {
		result.Routing = routingResult
	}

	// Step 4: 记录学习信号（任务开始）
	if e.collector != nil {
		e.collector.RecordImplicitFeedback(userID, "", "task_start", 0)
	}

	return result, nil
}

// RecordFeedback 记录用户反馈
func (e *Engine) RecordFeedback(userID int, taskID string, feedback *Feedback) {
	if e.collector == nil {
		return
	}

	switch feedback.Type {
	case "rating":
		e.collector.RecordExplicitFeedback(userID, taskID, feedback.Rating, feedback.Data)
	case "edit":
		e.collector.RecordEditAction(userID, taskID, feedback.PageIndex, feedback.Before, feedback.After)
	case "completion":
		e.collector.RecordTaskCompletion(userID, taskID, feedback.Duration, feedback.Rating)
	case "abandon":
		e.collector.RecordAbandonTask(userID, taskID, feedback.Reason, feedback.Progress)
	}
}

// UpdateProfileFromTask 从任务更新用户画像
func (e *Engine) UpdateProfileFromTask(userID int, task *TaskContext) {
	if e.updater == nil {
		return
	}

	// 提取风格
	extracted := &style.ExtractedStyle{
		Themes:   task.Themes,
		Colors:   task.Colors,
		PageCount: task.PageCount,
	}

	// 更新画像
	e.profileStore.UpdateWithTask(userID, extracted)

	// 更新领域偏好
	if task.Domain != "" {
		e.profileStore.UpdateDomainPreference(userID, task.Domain)
	}

	// 记录成功
	if task.Success {
		e.profileStore.RecordSuccess(userID, task.Domain, task.Template, task.Theme, task.PageCount)
	}

	// 记录学习信号
	if e.collector != nil {
		e.collector.RecordTaskCompletion(userID, task.TaskID, task.Duration, task.QualityScore)
	}
}

// GetRecommendations 获取个性化推荐
func (e *Engine) GetRecommendations(userID int, domain string) *style.RecommendResult {
	profile := e.profileStore.GetEnhanced(userID)
	if profile == nil {
		profile = style.NewEnhancedProfile(userID)
	}

	return profile.Recommend(&style.RecommendRequest{
		Domain:     domain,
		Complexity: 5,
	})
}

// GetUserInsights 获取用户洞察
func (e *Engine) GetUserInsights(userID int) *InsightsReport {
	if e.analyzer == nil {
		return nil
	}
	return e.analyzer.GenerateInsights(userID)
}

// GetProfile 获取用户画像
func (e *Engine) GetProfile(userID int) *style.EnhancedProfile {
	return e.profileStore.GetEnhanced(userID)
}

// Close 关闭引擎
func (e *Engine) Close() {
	if e.collector != nil {
		e.collector.Close()
	}
}

// ProcessingResult 处理结果
type ProcessingResult struct {
	UserID    int
	Query     string
	Timestamp time.Time
	Error     string

	Intent   *intent.ClassificationResult
	Routing  *intent.RoutingDecision
	Profile  *style.EnhancedProfile
}

// Feedback 用户反馈
type Feedback struct {
	Type      string
	Rating    float64
	PageIndex int
	Before    string
	After     string
	Duration  time.Duration
	Reason    string
	Progress  float64
	Data      map[string]interface{}
}

// TaskContext 任务上下文
type TaskContext struct {
	TaskID     string
	UserID    int
	Domain    string
	Template  string
	Theme     string
	PageCount int
	Duration  time.Duration
	Success   bool
	QualityScore float64
	Themes    []string
	Colors    []string
}

func now() time.Time {
	return time.Now()
}
