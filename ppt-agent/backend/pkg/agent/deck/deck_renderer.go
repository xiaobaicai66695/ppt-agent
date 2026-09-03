package deck

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino/compose"

	agentutils "github.com/cloudwego/ppt-agent/pkg/agent/utils"
	"github.com/cloudwego/ppt-agent/pkg/assets/unsplash"
	"github.com/cloudwego/ppt-agent/pkg/retry"
	"github.com/cloudwego/ppt-agent/pkg/tools/pythonutil"
)

type DeckRenderEvent struct {
	Type       string
	TaskID     string
	PageIndex  int
	OutputFile string
	Detail     string
	Error      string
}

type DeckRenderEventCallback func(event DeckRenderEvent)

type deckRenderInput struct {
	Config   *PPTTaskConfig
	Callback DeckRenderEventCallback
	Started  time.Time
}

type deckRenderContext struct {
	Config      *PPTTaskConfig
	Callback    DeckRenderEventCallback
	Started     time.Time
	Manifest    *TasksManifest
	Concurrency int
	Tasks       []*TaskItem
}

func RenderDeckByTaskIDWorkflow(ctx context.Context, cfg *PPTTaskConfig, onEvent DeckRenderEventCallback) (*PPTTaskResult, error) {
	graph := compose.NewGraph[*deckRenderInput, *PPTTaskResult]()
	_ = graph.AddLambdaNode("validate_deck_spec", compose.InvokableLambda(validateDeckRenderInput))
	_ = graph.AddLambdaNode("materialize_background_assets", compose.InvokableLambda(materializeDeckBackgroundAssets))
	_ = graph.AddLambdaNode("render_worker_pool", compose.InvokableLambda(renderValidatedDeck))
	_ = graph.AddEdge(compose.START, "validate_deck_spec")
	_ = graph.AddEdge("validate_deck_spec", "materialize_background_assets")
	_ = graph.AddEdge("materialize_background_assets", "render_worker_pool")
	_ = graph.AddEdge("render_worker_pool", compose.END)

	runner, err := graph.Compile(ctx, compose.WithGraphName("PPTDeckRenderWorkflow"))
	if err != nil {
		return nil, err
	}
	return runner.Invoke(ctx, &deckRenderInput{
		Config:   cfg,
		Callback: onEvent,
		Started:  time.Now(),
	})
}

func materializeDeckBackgroundAssets(ctx context.Context, deck *deckRenderContext) (*deckRenderContext, error) {
	if deck == nil || deck.Config == nil || deck.Manifest == nil || len(deck.Tasks) == 0 {
		return deck, nil
	}
	for revisionAttempt := 1; ; revisionAttempt++ {
		// Only hydrate pages selected for this render pass. The task pointers are
		// shared with deck.Manifest, so successful metadata is persisted below.
		pendingManifest := &TasksManifest{Tasks: deck.Tasks, VisualPolicy: deck.Manifest.VisualPolicy}
		counts, err := MaterializePlannedDeckAssets(ctx, deck.Config.WorkDir, pendingManifest)
		if errors.Is(err, unsplash.ErrMissingAccessKey) {
			return nil, fmt.Errorf("图片素材服务未配置，无法交付带背景图片的 PPT: %w", err)
		}
		if err != nil {
			var revision *assetQueryRevisionError
			if !errors.As(err, &revision) {
				return nil, fmt.Errorf("materialize planned deck assets: %w", err)
			}
			handled, revisionErr := backgroundSearchRetryFactory.Execute(ctx, retry.OperationUnsplashSearch, err, revisionAttempt, func(ctx context.Context, decision retry.Decision) error {
				return reviseBackgroundSearchTermsWithFixer(ctx, deck, revision, decision)
			})
			if revisionErr != nil {
				return nil, fmt.Errorf("背景图片搜索词 LLM 修订失败: %w", revisionErr)
			}
			if !handled {
				return nil, fmt.Errorf("背景图片搜索词经 LLM 修订后仍不可用: %w", err)
			}
			if err := reloadDeckRenderManifest(deck); err != nil {
				return nil, fmt.Errorf("读取 LLM 修订后的 DeckSpec 失败: %w", err)
			}
			continue
		}
		if missingPages := missingMaterializedBackgroundPages(deck.Config.WorkDir, pendingManifest); len(missingPages) > 0 {
			return nil, fmt.Errorf("背景图片未物化到本地，拒绝交付无背景 PPT：第 %s 页", formatPageIndexes(missingPages))
		}
		if counts.Backgrounds == 0 && counts.Images == 0 {
			return deck, nil
		}
		if err := WriteTasksManifest(deck.Config.WorkDir, deck.Manifest); err != nil {
			return nil, fmt.Errorf("persist materialized deck assets: %w", err)
		}
		if deck.Config.RuntimeMeta != nil {
			deck.Config.RuntimeMeta.RecordPhase(
				"compiling",
				fmt.Sprintf("已写回 %d 个轮换背景引用和 %d 个图文素材，开始按页渲染", counts.Backgrounds, counts.Images),
			)
		}
		return deck, nil
	}
}

func reviseBackgroundSearchTermsWithFixer(ctx context.Context, deck *deckRenderContext, revision *assetQueryRevisionError, decision retry.Decision) error {
	if deck == nil || deck.Config == nil || revision == nil {
		return fmt.Errorf("背景搜索词修订上下文不完整")
	}
	allowedTaskIDs, pageIndexes, queries := assetQueryRevisionTargets(revision.Requests)
	if len(allowedTaskIDs) == 0 {
		return fmt.Errorf("背景搜索词修订未包含可授权的任务 ID")
	}
	before, err := ReadTasksManifest(deck.Config.WorkDir)
	if err != nil {
		return err
	}
	beforeQueries := backgroundQuerySnapshot(before, allowedTaskIDs)
	emitRenderEvent(deck.Callback, DeckRenderEvent{Type: "asset_query_revision", Detail: fmt.Sprintf("背景素材服务拒绝第 %s 页搜索词，正在请求 LLM 重写", formatPageIndexes(pageIndexes))})
	if deck.Config.RuntimeMeta != nil {
		deck.Config.RuntimeMeta.RecordPhase("revising", fmt.Sprintf("背景素材 HTTP 410：按策略 %s 重写第 %s 页搜索词", decision.StrategyName, formatPageIndexes(pageIndexes)))
	}
	fixerCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	fixer, err := NewPPTFixerAgentForTasks(fixerCtx, deck.Config, allowedTaskIDs)
	if err != nil {
		return err
	}
	input := fmt.Sprintf("系统素材修订：Unsplash 背景图片搜索返回 HTTP 410，说明当前 asset_query 不适合素材服务。仅修改授权页面的背景 visual_intent.asset_query 或背景 image 组件的 asset_query。必须改成与原词不同、可搜索的简洁英文视觉主体词（2-5 个词）；不要重复原词、不要下载图片、不要改正文、标题、版式或其他字段。\n目标页面：%v\n失败搜索词：%v", pageIndexes, queries)
	if err := RunPPTFixerWithCallback(fixerCtx, fixer, input, func(event AgentEvent) {
		if event.Type == AgentEventProgress {
			emitRenderEvent(deck.Callback, DeckRenderEvent{Type: "asset_query_revision", Detail: event.PhaseDetail})
		}
	}); err != nil {
		return err
	}
	after, err := ReadTasksManifest(deck.Config.WorkDir)
	if err != nil {
		return err
	}
	if !backgroundQueriesChanged(beforeQueries, backgroundQuerySnapshot(after, allowedTaskIDs)) {
		return fmt.Errorf("LLM 未修改失败页的背景 asset_query")
	}
	return nil
}

func assetQueryRevisionTargets(requests []assetQueryRevisionRequest) ([]string, []int, []string) {
	taskSet, pageSet, querySet := map[string]struct{}{}, map[int]struct{}{}, map[string]struct{}{}
	for _, request := range requests {
		for _, taskID := range request.TaskIDs {
			taskSet[taskID] = struct{}{}
		}
		for _, pageIndex := range request.PageIndexes {
			pageSet[pageIndex] = struct{}{}
		}
		for _, query := range request.Queries {
			querySet[query] = struct{}{}
		}
	}
	taskIDs, pageIndexes, queries := make([]string, 0, len(taskSet)), make([]int, 0, len(pageSet)), make([]string, 0, len(querySet))
	for taskID := range taskSet {
		taskIDs = append(taskIDs, taskID)
	}
	for pageIndex := range pageSet {
		pageIndexes = append(pageIndexes, pageIndex)
	}
	for query := range querySet {
		queries = append(queries, query)
	}
	sort.Strings(taskIDs)
	sort.Ints(pageIndexes)
	sort.Strings(queries)
	return taskIDs, pageIndexes, queries
}

func backgroundQuerySnapshot(manifest *TasksManifest, taskIDs []string) map[string]string {
	allowed := map[string]struct{}{}
	for _, taskID := range taskIDs {
		allowed[taskID] = struct{}{}
	}
	result := map[string]string{}
	if manifest == nil {
		return result
	}
	for _, task := range manifest.Tasks {
		if task == nil {
			continue
		}
		if _, ok := allowed[task.TaskID]; !ok {
			continue
		}
		if task.ContentPlan == nil {
			continue
		}
		queries := []string{}
		if visual := task.ContentPlan.VisualIntent; visual != nil && isBackgroundPlan(visual.AssetPurpose, visual.ImagePosition, visual.Role) {
			queries = append(queries, strings.TrimSpace(visual.AssetQuery))
		}
		for index := range task.ContentPlan.Components {
			component := &task.ContentPlan.Components[index]
			if component.Type == "image" && isBackgroundPlan(component.AssetPurpose, "", component.Role) {
				queries = append(queries, strings.TrimSpace(component.AssetQuery))
			}
		}
		result[task.TaskID] = strings.Join(queries, "\x00")
	}
	return result
}

func backgroundQueriesChanged(before, after map[string]string) bool {
	for taskID, beforeQuery := range before {
		if afterQuery := strings.TrimSpace(after[taskID]); afterQuery != "" && afterQuery != beforeQuery {
			return true
		}
	}
	return false
}

func reloadDeckRenderManifest(deck *deckRenderContext) error {
	selected := map[string]struct{}{}
	for _, task := range deck.Tasks {
		if task != nil {
			selected[task.TaskID] = struct{}{}
		}
	}
	manifest, err := ReadTasksManifest(deck.Config.WorkDir)
	if err != nil {
		return err
	}
	tasks := make([]*TaskItem, 0, len(selected))
	for _, task := range manifest.Tasks {
		if task != nil {
			if _, ok := selected[task.TaskID]; ok {
				tasks = append(tasks, task)
			}
		}
	}
	if len(tasks) != len(selected) {
		return fmt.Errorf("修订后 DeckSpec 缺少待渲染页面")
	}
	deck.Manifest, deck.Tasks = manifest, tasks
	return nil
}

func formatPageIndexes(pageIndexes []int) string {
	values := make([]string, 0, len(pageIndexes))
	for _, pageIndex := range pageIndexes {
		values = append(values, strconv.Itoa(pageIndex))
	}
	return strings.Join(values, ",")
}

func validateDeckRenderInput(ctx context.Context, input *deckRenderInput) (*deckRenderContext, error) {
	_ = ctx
	if input == nil {
		return nil, fmt.Errorf("nil deck render input")
	}
	cfg := input.Config
	if cfg == nil {
		return nil, fmt.Errorf("nil PPTTaskConfig")
	}
	if strings.TrimSpace(cfg.WorkDir) == "" {
		return nil, fmt.Errorf("workDir is empty")
	}
	if strings.TrimSpace(cfg.SkillsDir) == "" {
		return nil, fmt.Errorf("skillsDir is empty")
	}

	manifest, err := ReadTasksManifest(cfg.WorkDir)
	if err != nil {
		return nil, fmt.Errorf("read tasks manifest before rendering: %w", err)
	}
	if manifest == nil || len(manifest.Tasks) == 0 {
		return nil, fmt.Errorf("tasks.json has no slides to render")
	}
	if err := validateManifestForWrite(manifest); err != nil {
		return nil, fmt.Errorf("DeckSpec validation failed: %w", err)
	}
	if cfg.RuntimeMeta != nil {
		cfg.RuntimeMeta.RecordPhase("compiling", "校验 DeckSpec 并准备按页渲染")
		cfg.RuntimeMeta.FreezePlan(renderPlanSlides(manifest.Tasks))
	}

	tasks := tasksNeedingRender(cfg.WorkDir, manifest.Tasks)
	if len(tasks) == 0 {
		reconciled, recErr := ReconcileTasksManifestOutputFiles(cfg.WorkDir)
		if recErr != nil {
			return nil, recErr
		}
		return &deckRenderContext{
			Config:   cfg,
			Callback: input.Callback,
			Started:  input.Started,
			Manifest: reconciled,
		}, nil
	}

	concurrency := 5
	if cfg.Concurrency > 0 {
		concurrency = cfg.Concurrency
	}
	if concurrency < 1 {
		concurrency = 1
	}
	if concurrency > len(tasks) {
		concurrency = len(tasks)
	}

	return &deckRenderContext{
		Config:      cfg,
		Callback:    input.Callback,
		Started:     input.Started,
		Manifest:    manifest,
		Concurrency: concurrency,
		Tasks:       tasks,
	}, nil
}

func renderValidatedDeck(ctx context.Context, deck *deckRenderContext) (*PPTTaskResult, error) {
	if deck == nil || deck.Config == nil {
		return nil, fmt.Errorf("nil deck render context")
	}
	cfg := deck.Config
	if len(deck.Tasks) == 0 {
		return renderResultFromManifest(cfg.WorkDir, deck.Manifest, deck.Started), nil
	}

	if cfg.RuntimeMeta != nil {
		cfg.RuntimeMeta.RecordPhase("rendering", fmt.Sprintf("按 task_id 并发渲染 %d 页，并发数 %d", len(deck.Tasks), deck.Concurrency))
	}
	emitRenderEvent(deck.Callback, DeckRenderEvent{Type: "workflow_start", Detail: fmt.Sprintf("并发渲染 %d 页", len(deck.Tasks))})

	jobs := make(chan *TaskItem)
	errs := make(chan error, len(deck.Tasks))
	var wg sync.WaitGroup
	for i := 0; i < deck.Concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for task := range jobs {
				if task == nil {
					continue
				}
				if err := renderOneTask(ctx, cfg, task, deck.Callback); err != nil {
					errs <- err
				}
			}
		}(i + 1)
	}
	for _, task := range deck.Tasks {
		select {
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			return nil, ctx.Err()
		case jobs <- task:
		}
	}
	close(jobs)
	wg.Wait()
	close(errs)

	var failures []string
	for err := range errs {
		if err != nil {
			failures = append(failures, err.Error())
		}
	}

	reconciled, recErr := ReconcileTasksManifestOutputFiles(cfg.WorkDir)
	if recErr != nil {
		return nil, recErr
	}
	result := renderResultFromManifest(cfg.WorkDir, reconciled, deck.Started)
	if len(failures) > 0 {
		if cfg.RuntimeMeta != nil {
			cfg.RuntimeMeta.RecordPhase("render_failed", fmt.Sprintf("%d 页渲染失败", len(failures)))
		}
		return result, fmt.Errorf("render workflow failed: %s", strings.Join(failures, " | "))
	}
	if cfg.RuntimeMeta != nil {
		cfg.RuntimeMeta.RecordPhase("rendered", fmt.Sprintf("完成 %d/%d 页渲染", result.DoneSlides, result.TotalSlides))
	}
	return result, nil
}

func tasksNeedingRender(workDir string, tasks []*TaskItem) []*TaskItem {
	var result []*TaskItem
	for _, task := range tasks {
		if task == nil {
			continue
		}
		switch strings.TrimSpace(task.Status) {
		case StatusPending, StatusGenerating, StatusFailed:
			result = append(result, task)
			continue
		}
		if task.OutputFile != "" {
			if _, err := os.Stat(filepath.Join(workDir, task.OutputFile)); err == nil {
				continue
			}
		}
		if task.Status == StatusDone || task.Status == StatusQADone || task.Status == StatusFixed {
			continue
		}
		result = append(result, task)
	}
	return result
}

func renderOneTask(ctx context.Context, cfg *PPTTaskConfig, task *TaskItem, onEvent DeckRenderEventCallback) error {
	taskID := strings.TrimSpace(task.TaskID)
	if taskID == "" {
		taskID = fmt.Sprint(task.PageIndex)
	}
	emitRenderEvent(onEvent, DeckRenderEvent{
		Type:       "slide_start",
		TaskID:     taskID,
		PageIndex:  task.PageIndex,
		OutputFile: task.OutputFile,
		Detail:     task.Title,
	})
	if cfg.RuntimeMeta != nil {
		cfg.RuntimeMeta.RecordToolStart("generate_slide", fmt.Sprintf(`{"task_id":"%s"}`, taskID))
	}
	if err := PatchTaskRuntimeFields(cfg.WorkDir, taskID, StatusGenerating); err != nil {
		return err
	}

	output, err := runRenderTaskScript(ctx, cfg, taskID)
	if err != nil {
		msg := trimCommandOutput(output)
		if msg == "" {
			msg = err.Error()
		}
		patchErr := PatchTaskRuntimeFields(cfg.WorkDir, taskID, StatusFailed)
		if cfg.RuntimeMeta != nil {
			cfg.RuntimeMeta.RecordToolErrorDetails("generate_slide", msg, map[string]any{
				"task_id": taskID,
				"output":  msg,
			})
		}
		emitRenderEvent(onEvent, DeckRenderEvent{
			Type:       "slide_error",
			TaskID:     taskID,
			PageIndex:  task.PageIndex,
			OutputFile: task.OutputFile,
			Error:      msg,
		})
		if patchErr != nil {
			return fmt.Errorf("task %s render failed: %s; status patch failed: %w", taskID, msg, patchErr)
		}
		return fmt.Errorf("task %s render failed: %s", taskID, msg)
	}

	if task.OutputFile != "" {
		outputPath := filepath.Join(cfg.WorkDir, task.OutputFile)
		if _, statErr := os.Stat(outputPath); statErr != nil {
			msg := fmt.Sprintf("expected output file missing: %s", task.OutputFile)
			_ = PatchTaskRuntimeFields(cfg.WorkDir, taskID, StatusFailed)
			return fmt.Errorf("task %s render failed: %s", taskID, msg)
		}
		if cfg.RuntimeMeta != nil {
			cfg.RuntimeMeta.RecordFileCreated(task.OutputFile)
		}
	}
	if err := PatchTaskRuntimeFields(cfg.WorkDir, taskID, StatusDone); err != nil {
		return err
	}
	if cfg.RuntimeMeta != nil {
		cfg.RuntimeMeta.RecordToolEnd("generate_slide", fmt.Sprintf(`{"task_id":"%s"}`, taskID), trimCommandOutput(output))
	}
	emitRenderEvent(onEvent, DeckRenderEvent{
		Type:       "slide_done",
		TaskID:     taskID,
		PageIndex:  task.PageIndex,
		OutputFile: task.OutputFile,
		Detail:     trimCommandOutput(output),
	})
	return nil
}

func runRenderTaskScript(ctx context.Context, cfg *PPTTaskConfig, taskID string) (string, error) {
	scriptPath := filepath.Join(cfg.SkillsDir, "ppt-deck-planner", "generators", "render_task.py")
	timeout := time.Duration(agentRenderTimeoutSeconds()) * time.Second
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(execCtx, pythonutil.GetPythonBinary(),
		scriptPath,
		"--work-dir", cfg.WorkDir,
		"--skills-dir", cfg.SkillsDir,
		"--task-id", taskID,
	)
	cmd.Env = append(os.Environ(), "PYTHONIOENCODING=utf-8")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if execCtx.Err() == context.DeadlineExceeded {
		return out.String(), fmt.Errorf("render task timed out after %s", timeout)
	}
	return out.String(), err
}

func agentRenderTimeoutSeconds() int {
	if value := strings.TrimSpace(os.Getenv("SLIDE_RENDER_TIMEOUT_SECONDS")); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 {
			return parsed
		}
	}
	return 180
}

func renderResultFromManifest(workDir string, manifest *TasksManifest, started time.Time) *PPTTaskResult {
	result := &PPTTaskResult{Duration: time.Since(started)}
	if manifest == nil {
		return result
	}
	result.TotalSlides = len(manifest.Tasks)
	result.DoneSlides = manifest.CompletedCount()
	for _, task := range manifest.Tasks {
		if task == nil || task.OutputFile == "" {
			continue
		}
		if task.Status == StatusDone || task.Status == StatusQADone || task.Status == StatusFixed {
			result.Files = append(result.Files, filepath.Join(workDir, task.OutputFile))
		}
	}
	return result
}

func emitRenderEvent(callback DeckRenderEventCallback, event DeckRenderEvent) {
	if callback != nil {
		callback(event)
	}
}

func trimCommandOutput(output string) string {
	output = strings.TrimSpace(output)
	if len(output) <= 4000 {
		return output
	}
	return output[:4000] + "...(truncated)"
}

func renderPlanSlides(items []*TaskItem) []agentutils.PlanSlide {
	slides := make([]agentutils.PlanSlide, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		slides = append(slides, agentutils.PlanSlide{
			PageIndex:   item.PageIndex,
			TaskID:      item.TaskID,
			Title:       item.Title,
			ContentType: item.ContentType,
			OutputFile:  item.OutputFile,
			Status:      item.Status,
		})
	}
	return slides
}
