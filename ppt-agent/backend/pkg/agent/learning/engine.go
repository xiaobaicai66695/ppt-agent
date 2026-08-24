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
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/cloudwego/ppt-agent/pkg/agent/intent"
	"github.com/cloudwego/ppt-agent/pkg/agent/router"
	"github.com/cloudwego/ppt-agent/pkg/logger"
	"github.com/cloudwego/ppt-agent/pkg/style"
)

// Engine 承载意图识别、路由和只读画像推荐。
// 历史包名保留为 learning，但自动学习、反馈分析和任务完成写画像已停用。
type Engine struct {
	classifier       *intent.Classifier
	router           *router.Engine
	profileStore     *style.EnhancedProfileStore
	textModelFactory interface{}
}

// EngineConfig 引擎配置
type EngineConfig struct {
	EnableLLMClassification bool // 是否启用LLM意图分类
	EnableLearning          bool // deprecated: 自动学习已停用，仅保留字段兼容旧配置
	EnableProfileMatch      bool // 是否启用画像匹配
}

// NewEngine 创建智能学习引擎
// modelFactory 用于创建 LLM 实例（用于意图分类等辅助任务）
// textModelFactory 用于创建轻量级 LLM 实例（优先使用，节省成本）
// 路由分类只需要 Generate 能力；只有明确传入 ToolCallingChatModel 工厂时，
// 才启用工具模型兜底，避免把 Web 层的 Generate-only adapter 误转为 ToolCalling。
// 如果传入 nil，则意图分类器只使用规则匹配
func NewEngine(cfg *EngineConfig, modelFactory interface{}, textModelFactory interface{}) *Engine {
	if cfg == nil {
		cfg = &EngineConfig{
			EnableLLMClassification: true,
			EnableLearning:          false,
			EnableProfileMatch:      true,
		}
	}

	e := &Engine{}

	// 初始化增强画像存储（先于其他组件）
	e.profileStore = style.NewEnhancedProfileStore()
	e.textModelFactory = textModelFactory

	// 初始化意图分类器（可选择是否启用 LLM）
	routingTextFactory := e.makeTextModelFactory(textModelFactory)
	if routingTextFactory == nil {
		routingTextFactory = e.makeTextModelFactory(modelFactory)
	}
	e.classifier = intent.NewClassifier(e.makeClassifierFactory(modelFactory), routingTextFactory)

	// 初始化路由引擎
	e.router = router.NewEngine(e.classifier)

	logger.Info("intelligent_engine_initialized",
		"llm_classification", cfg.EnableLLMClassification,
		"learning", false,
		"profile_match", cfg.EnableProfileMatch,
		"llm_enabled", hasCallableFactory(modelFactory))

	return e
}

// makeClassifierFactory 将通用 modelFactory 适配为 intent.Classifier 需要的类型
func (e *Engine) makeClassifierFactory(modelFactory interface{}) func(ctx context.Context) (model.ToolCallingChatModel, error) {
	if !hasCallableFactory(modelFactory) {
		return nil
	}

	switch f := modelFactory.(type) {
	case func(ctx context.Context) (model.ToolCallingChatModel, error):
		if reflect.ValueOf(f).IsNil() {
			return nil
		}
		return f
	default:
		return nil
	}
}

// callModelFactory 动态调用 modelFactory（支持多种函数签名）
func callModelFactory(modelFactory interface{}, ctx context.Context) (interface{}, error) {
	if !hasCallableFactory(modelFactory) {
		return nil, fmt.Errorf("modelFactory 为空")
	}
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

// makeTextModelFactory 将通用 textModelFactory 适配为 intent.Classifier 需要的类型
func (e *Engine) makeTextModelFactory(factory interface{}) func(ctx context.Context) (interface {
	Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (msg *schema.Message, err error)
}, error) {
	if !hasCallableFactory(factory) {
		return nil
	}

	return func(ctx context.Context) (interface {
		Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (msg *schema.Message, err error)
	}, error) {
		result, err := callModelFactory(factory, ctx)
		if err != nil {
			return nil, err
		}
		if gm, ok := result.(interface {
			Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (msg *schema.Message, err error)
		}); ok {
			return gm, nil
		}
		return nil, fmt.Errorf("textModelFactory 返回值不满足 Generate 接口: %T", result)
	}
}

func hasCallableFactory(factory interface{}) bool {
	if factory == nil {
		return false
	}
	v := reflect.ValueOf(factory)
	if !v.IsValid() || v.Kind() != reflect.Func {
		return false
	}
	return !v.IsNil()
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

	// Step 2: 获取用户画像。未登录的公开接口没有用户画像，不能访问 DB。
	profile := style.NewEnhancedProfile(userID)
	if userID > 0 {
		profile = e.profileStore.GetEnhanced(userID)
	}
	result.Profile = profile

	// Step 3: reuse the same LLM classification for routing. Routing must not
	// invoke a second model call or semantic rule pass.
	routingResult, err := e.router.Route(intentResult, profile)
	if err != nil {
		logger.Warn("routing_failed", "error", err.Error())
		result.Error = err.Error()
	} else {
		result.Routing = routingResult
	}

	return result, nil
}

// GetRecommendations 获取只读画像推荐
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

// GetProfile 获取用户画像
func (e *Engine) GetProfile(userID int) *style.EnhancedProfile {
	return e.profileStore.GetEnhanced(userID)
}

// Close 关闭引擎
func (e *Engine) Close() {}

// ProcessingResult 处理结果
type ProcessingResult struct {
	UserID    int
	Query     string
	Timestamp time.Time
	Error     string

	Intent  *intent.ClassificationResult
	Routing *intent.RoutingDecision
	Profile *style.EnhancedProfile
}

func now() time.Time {
	return time.Now()
}
