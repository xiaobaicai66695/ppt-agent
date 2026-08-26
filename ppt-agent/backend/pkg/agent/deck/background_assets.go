package deck

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/cloudwego/ppt-agent/pkg/assets/unsplash"
	"github.com/cloudwego/ppt-agent/pkg/logger"
)

const (
	maxBackgroundAssetWorkers = 3
)

type backgroundAssetClient interface {
	Search(context.Context, unsplash.SearchOptions) (*unsplash.SearchResponse, error)
	Download(context.Context, unsplash.Photo, string) (*unsplash.DownloadedAsset, error)
}

type plannedBackgroundTarget struct {
	query       string
	subject     string
	contentType string
	slot        int
	searchPage  int
	visual      *VisualIntent
	component   *PlanComponent
}

type resolvedBackgroundAsset struct {
	photo unsplash.Photo
	asset *unsplash.DownloadedAsset
}

type plannedImageAssetTarget struct {
	query     string
	component *PlanComponent
}

type MaterializedDeckAssetCounts struct {
	Backgrounds int
	Images      int
}

// MaterializePlannedDeckAssets resolves all query-only external image plans
// into downloaded task-local files. Backgrounds are intentionally collapsed by
// slide content_type so pages of the same type share one visual rhythm;
// foreground image components keep their page-specific queries.
func MaterializePlannedDeckAssets(ctx context.Context, workDir string, manifest *TasksManifest) (MaterializedDeckAssetCounts, error) {
	backgroundTargets, err := collectDeckBackgroundTargets(workDir, manifest)
	if err != nil {
		return MaterializedDeckAssetCounts{}, err
	}
	imageTargets, err := collectPendingImageAssetTargets(workDir, manifest)
	if err != nil {
		return MaterializedDeckAssetCounts{}, err
	}
	if len(backgroundTargets) == 0 && len(imageTargets) == 0 {
		return MaterializedDeckAssetCounts{}, nil
	}
	var client backgroundAssetClient
	if needsAssetClient(workDir, backgroundTargets, imageTargets) {
		var err error
		client, err = unsplash.NewClientFromEnv()
		if err != nil {
			return MaterializedDeckAssetCounts{}, err
		}
	}
	backgroundCount, err := materializePlannedBackgroundsWithClient(ctx, workDir, backgroundTargets, client)
	if err != nil {
		return MaterializedDeckAssetCounts{}, err
	}
	imageCount, err := materializePlannedImageAssetsWithClient(ctx, workDir, imageTargets, client)
	if err != nil {
		return MaterializedDeckAssetCounts{}, err
	}
	return MaterializedDeckAssetCounts{Backgrounds: backgroundCount, Images: imageCount}, nil
}

func needsAssetClient(workDir string, backgroundTargets []plannedBackgroundTarget, imageTargets []plannedImageAssetTarget) bool {
	if len(imageTargets) > 0 {
		return true
	}
	absWorkDir, err := filepath.Abs(strings.TrimSpace(workDir))
	if err != nil {
		return true
	}
	assigned := assignDeckBackgroundRotation(append([]plannedBackgroundTarget(nil), backgroundTargets...))
	slots := make(map[int]bool)
	for _, target := range assigned {
		if _, ok := slots[target.slot]; ok {
			continue
		}
		if _, ok := resolvedExistingBackgroundAsset(absWorkDir, target); ok {
			slots[target.slot] = true
		}
	}
	for _, target := range assigned {
		if !slots[target.slot] {
			return true
		}
	}
	return false
}

// MaterializePlannedBackgrounds resolves query-only background plans into
// downloaded task-local files. The caller persists the updated manifest only
// after every required search/download succeeds, avoiding partial DeckSpecs.
func MaterializePlannedBackgrounds(ctx context.Context, workDir string, manifest *TasksManifest) (int, error) {
	targets, err := collectPendingBackgroundTargets(workDir, manifest)
	if err != nil || len(targets) == 0 {
		return 0, err
	}
	client, err := unsplash.NewClientFromEnv()
	if err != nil {
		return 0, err
	}
	return materializePlannedBackgroundsWithClient(ctx, workDir, targets, client)
}

func collectPendingBackgroundTargets(workDir string, manifest *TasksManifest) ([]plannedBackgroundTarget, error) {
	if manifest == nil {
		return nil, fmt.Errorf("nil tasks manifest")
	}
	workDir = strings.TrimSpace(workDir)
	if workDir == "" {
		return nil, fmt.Errorf("background asset work directory is empty")
	}
	absWorkDir, err := filepath.Abs(workDir)
	if err != nil {
		return nil, fmt.Errorf("resolve background asset work directory: %w", err)
	}

	var targets []plannedBackgroundTarget
	for _, task := range manifest.Tasks {
		if task == nil || task.ContentPlan == nil {
			continue
		}
		if visual := task.ContentPlan.VisualIntent; visual != nil && isBackgroundPlan(visual.AssetPurpose, visual.ImagePosition, visual.Role) {
			if !taskLocalAssetExists(absWorkDir, visual.LocalPath) && strings.TrimSpace(visual.AssetQuery) != "" {
				targets = append(targets, plannedBackgroundTarget{
					query:       strings.TrimSpace(visual.AssetQuery),
					subject:     strings.TrimSpace(visual.AssetSubject),
					contentType: strings.TrimSpace(task.ContentType),
					searchPage:  1,
					visual:      visual,
				})
			}
		}
		for index := range task.ContentPlan.Components {
			component := &task.ContentPlan.Components[index]
			if component.Type != "image" || !isBackgroundPlan(component.AssetPurpose, "", component.Role) {
				continue
			}
			if !taskLocalAssetExists(absWorkDir, component.LocalPath) && strings.TrimSpace(component.AssetQuery) != "" {
				targets = append(targets, plannedBackgroundTarget{
					query:       strings.TrimSpace(component.AssetQuery),
					subject:     strings.TrimSpace(component.AssetSubject),
					contentType: strings.TrimSpace(task.ContentType),
					searchPage:  1,
					component:   component,
				})
			}
		}
	}
	return targets, nil
}

func collectDeckBackgroundTargets(workDir string, manifest *TasksManifest) ([]plannedBackgroundTarget, error) {
	if manifest == nil {
		return nil, fmt.Errorf("nil tasks manifest")
	}
	workDir = strings.TrimSpace(workDir)
	if workDir == "" {
		return nil, fmt.Errorf("background asset work directory is empty")
	}
	absWorkDir, err := filepath.Abs(workDir)
	if err != nil {
		return nil, fmt.Errorf("resolve background asset work directory: %w", err)
	}
	var targets []plannedBackgroundTarget
	for _, task := range manifest.Tasks {
		if task == nil || task.ContentPlan == nil {
			continue
		}
		if visual := task.ContentPlan.VisualIntent; visual != nil && isBackgroundPlan(visual.AssetPurpose, visual.ImagePosition, visual.Role) {
			if taskLocalAssetExists(absWorkDir, visual.LocalPath) || strings.TrimSpace(visual.AssetQuery) != "" {
				targets = append(targets, plannedBackgroundTarget{
					query:       strings.TrimSpace(visual.AssetQuery),
					subject:     strings.TrimSpace(visual.AssetSubject),
					contentType: strings.TrimSpace(task.ContentType),
					searchPage:  1,
					visual:      visual,
				})
			}
		}
		for index := range task.ContentPlan.Components {
			component := &task.ContentPlan.Components[index]
			if component.Type != "image" || !isBackgroundPlan(component.AssetPurpose, "", component.Role) {
				continue
			}
			if taskLocalAssetExists(absWorkDir, component.LocalPath) || strings.TrimSpace(component.AssetQuery) != "" {
				targets = append(targets, plannedBackgroundTarget{
					query:       strings.TrimSpace(component.AssetQuery),
					subject:     strings.TrimSpace(component.AssetSubject),
					contentType: strings.TrimSpace(task.ContentType),
					searchPage:  1,
					component:   component,
				})
			}
		}
	}
	return targets, nil
}

func materializePlannedBackgroundsWithClient(ctx context.Context, workDir string, targets []plannedBackgroundTarget, client backgroundAssetClient) (int, error) {
	if len(targets) == 0 {
		return 0, nil
	}
	absWorkDir, err := filepath.Abs(workDir)
	if err != nil {
		return 0, err
	}
	downloadDir := filepath.Join(absWorkDir, "assets", "images")
	targets = assignDeckBackgroundRotation(targets)

	slotAssets := make(map[int]resolvedBackgroundAsset)
	for _, target := range targets {
		if _, ok := slotAssets[target.slot]; ok {
			continue
		}
		if resolved, ok := resolvedExistingBackgroundAsset(absWorkDir, target); ok {
			slotAssets[target.slot] = resolved
		}
	}

	groups := make(map[string][]plannedBackgroundTarget)
	for _, target := range targets {
		if _, ok := slotAssets[target.slot]; ok {
			continue
		}
		key := backgroundTargetKey(target)
		if key == "" {
			continue
		}
		groups[key] = append(groups[key], target)
	}
	keys := make([]string, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	resolved := make([]resolvedBackgroundAsset, len(keys))
	errs := make([]error, len(keys))
	if len(keys) > 0 && client == nil {
		return 0, fmt.Errorf("background image client is nil")
	}
	jobs := make(chan int)
	workerCount := maxBackgroundAssetWorkers
	if workerCount > len(keys) {
		workerCount = len(keys)
	}
	var wg sync.WaitGroup
	for worker := 0; worker < workerCount; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for index := range jobs {
				target := groups[keys[index]][0]
				query := target.query
				result, searchErr := client.Search(ctx, unsplash.SearchOptions{
					Query:         query,
					Orientation:   "landscape",
					ContentFilter: "high",
					OrderBy:       "relevant",
					Page:          target.searchPage,
					PerPage:       1,
				})
				if searchErr != nil {
					errs[index] = fmt.Errorf("search background for %q: %w", query, searchErr)
					continue
				}
				if result == nil || len(result.Results) == 0 {
					errs[index] = fmt.Errorf("search background for %q returned no photos", query)
					continue
				}
				photo := result.Results[0]
				asset, downloadErr := client.Download(ctx, photo, downloadDir)
				if downloadErr != nil {
					errs[index] = fmt.Errorf("download background for %q: %w", query, downloadErr)
					continue
				}
				if asset == nil || !taskLocalAssetExists(absWorkDir, asset.LocalPath) {
					errs[index] = fmt.Errorf("download background for %q did not create a task-local file", query)
					continue
				}
				resolved[index] = resolvedBackgroundAsset{photo: photo, asset: asset}
			}
		}()
	}
	for index := range keys {
		select {
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			return 0, ctx.Err()
		case jobs <- index:
		}
	}
	close(jobs)
	wg.Wait()

	for _, err := range errs {
		if err != nil {
			if isRecoverableAssetError(err) {
				logger.Warn("background_asset_skipped", "error", err.Error())
				continue
			}
			return 0, err
		}
	}
	for index, key := range keys {
		if len(groups[key]) > 0 && resolved[index].asset != nil {
			slotAssets[groups[key][0].slot] = resolved[index]
		}
	}
	count := 0
	for _, target := range targets {
		if resolved, ok := slotAssets[target.slot]; ok {
			applyResolvedBackground(target, resolved)
			count++
		}
	}
	return count, nil
}

func collectPendingImageAssetTargets(workDir string, manifest *TasksManifest) ([]plannedImageAssetTarget, error) {
	if manifest == nil {
		return nil, fmt.Errorf("nil tasks manifest")
	}
	workDir = strings.TrimSpace(workDir)
	if workDir == "" {
		return nil, fmt.Errorf("image asset work directory is empty")
	}
	absWorkDir, err := filepath.Abs(workDir)
	if err != nil {
		return nil, fmt.Errorf("resolve image asset work directory: %w", err)
	}
	var targets []plannedImageAssetTarget
	for _, task := range manifest.Tasks {
		if task == nil || task.ContentPlan == nil {
			continue
		}
		for index := range task.ContentPlan.Components {
			component := &task.ContentPlan.Components[index]
			if component.Type != "image" || isBackgroundPlan(component.AssetPurpose, "", component.Role) {
				continue
			}
			if !taskLocalAssetExists(absWorkDir, component.LocalPath) && strings.TrimSpace(component.AssetQuery) != "" {
				targets = append(targets, plannedImageAssetTarget{
					query:     normalizeAssetQuery(strings.TrimSpace(component.AssetQuery), false),
					component: component,
				})
			}
		}
	}
	return targets, nil
}

func materializePlannedImageAssetsWithClient(ctx context.Context, workDir string, targets []plannedImageAssetTarget, client backgroundAssetClient) (int, error) {
	if len(targets) == 0 {
		return 0, nil
	}
	if client == nil {
		return 0, fmt.Errorf("image asset client is nil")
	}
	absWorkDir, err := filepath.Abs(workDir)
	if err != nil {
		return 0, err
	}
	downloadDir := filepath.Join(absWorkDir, "assets", "images")
	groups := make(map[string][]plannedImageAssetTarget)
	for _, target := range targets {
		key := strings.ToLower(strings.TrimSpace(target.query))
		if key == "" {
			continue
		}
		groups[key] = append(groups[key], target)
	}
	keys := make([]string, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	resolved := make([]resolvedBackgroundAsset, len(keys))
	errs := make([]error, len(keys))
	jobs := make(chan int)
	workerCount := maxBackgroundAssetWorkers
	if workerCount > len(keys) {
		workerCount = len(keys)
	}
	var wg sync.WaitGroup
	for worker := 0; worker < workerCount; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for index := range jobs {
				query := groups[keys[index]][0].query
				result, searchErr := client.Search(ctx, unsplash.SearchOptions{
					Query:         query,
					Orientation:   "landscape",
					ContentFilter: "high",
					OrderBy:       "relevant",
					Page:          1,
					PerPage:       1,
				})
				if searchErr != nil {
					errs[index] = fmt.Errorf("search image asset for %q: %w", query, searchErr)
					continue
				}
				if result == nil || len(result.Results) == 0 {
					errs[index] = fmt.Errorf("search image asset for %q returned no photos", query)
					continue
				}
				photo := result.Results[0]
				asset, downloadErr := client.Download(ctx, photo, downloadDir)
				if downloadErr != nil {
					errs[index] = fmt.Errorf("download image asset for %q: %w", query, downloadErr)
					continue
				}
				if asset == nil || !taskLocalAssetExists(absWorkDir, asset.LocalPath) {
					errs[index] = fmt.Errorf("download image asset for %q did not create a task-local file", query)
					continue
				}
				resolved[index] = resolvedBackgroundAsset{photo: photo, asset: asset}
			}
		}()
	}
	for index := range keys {
		select {
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			return 0, ctx.Err()
		case jobs <- index:
		}
	}
	close(jobs)
	wg.Wait()
	for _, err := range errs {
		if err != nil {
			if isRecoverableAssetError(err) {
				logger.Warn("image_asset_skipped", "error", err.Error())
				continue
			}
			return 0, err
		}
	}
	count := 0
	for index, key := range keys {
		if resolved[index].asset == nil {
			continue
		}
		for _, target := range groups[key] {
			applyResolvedImageAsset(target, resolved[index])
			count++
		}
	}
	return count, nil
}

func isRecoverableAssetError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) ||
		errors.Is(err, unsplash.ErrMissingAccessKey) || errors.Is(err, unsplash.ErrUnauthorized) {
		return false
	}
	message := strings.ToLower(err.Error())
	if strings.Contains(message, "http 401") {
		return false
	}
	recoverableSnippets := []string{
		"returned no photos",
		"content removed",
		"http 404",
		"http 410",
		"http 429",
		"http 500",
		"http 502",
		"http 503",
		"http 504",
		"call unsplash search api",
		"download unsplash image",
		"track unsplash download",
	}
	for _, snippet := range recoverableSnippets {
		if strings.Contains(message, snippet) {
			return true
		}
	}
	return false
}

func assignDeckBackgroundRotation(targets []plannedBackgroundTarget) []plannedBackgroundTarget {
	if len(targets) == 0 {
		return targets
	}
	typeSlots := map[string]int{}
	typeQueries := map[string]string{}
	for _, target := range targets {
		typeKey := backgroundContentTypeKey(target.contentType)
		if _, ok := typeSlots[typeKey]; !ok {
			typeSlots[typeKey] = len(typeSlots)
		}
		if typeQueries[typeKey] == "" {
			typeQueries[typeKey] = normalizeAssetQuery(firstNonBlank(target.subject, target.query), true)
		}
	}
	for index := range targets {
		typeKey := backgroundContentTypeKey(targets[index].contentType)
		slot := typeSlots[typeKey]
		targets[index].slot = slot
		if query := typeQueries[typeKey]; query != "" {
			targets[index].query = query
		}
		targets[index].searchPage = 1
	}
	return targets
}

func backgroundContentTypeKey(contentType string) string {
	if key := strings.ToLower(strings.TrimSpace(contentType)); key != "" {
		return key
	}
	return "default"
}

func backgroundTargetKey(target plannedBackgroundTarget) string {
	query := strings.ToLower(strings.TrimSpace(target.query))
	if query == "" {
		return ""
	}
	return fmt.Sprintf("%02d:%02d:%s", target.slot, target.searchPage, query)
}

func normalizeAssetQuery(query string, background bool) string {
	query = strings.Join(strings.Fields(strings.TrimSpace(query)), " ")
	if query == "" || !background {
		return query
	}
	return compactBackgroundSearchKeyword(query)
}

func compactBackgroundSearchKeyword(query string) string {
	query = strings.TrimSpace(strings.ToLower(query))
	query = strings.NewReplacer(
		",", " ", ".", " ", ";", " ", ":", " ", "/", " ", "\\", " ", "|", " ",
		"-", " ", "_", " ", "(", " ", ")", " ", "[", " ", "]", " ",
	).Replace(query)
	words := strings.Fields(query)
	if len(words) == 0 {
		return ""
	}
	stop := map[string]bool{
		"a": true, "an": true, "the": true, "and": true, "or": true, "of": true, "for": true, "with": true,
		"light": true, "bright": true, "pale": true, "white": true, "airy": true, "clean": true,
		"wide": true, "landscape": true, "background": true, "negative": true, "space": true,
		"photo": true, "image": true, "abstract": true, "global": true, "modern": true, "theme": true,
		"slide": true, "deck": true, "presentation": true,
	}
	for _, word := range words {
		word = strings.Trim(word, "'\"")
		if word == "" || stop[word] {
			continue
		}
		if containsDigitOnly(word) {
			continue
		}
		return word
	}
	return words[0]
}

func containsDigitOnly(value string) bool {
	if value == "" {
		return false
	}
	for _, ch := range value {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

func applyResolvedBackground(target plannedBackgroundTarget, resolved resolvedBackgroundAsset) {
	asset := resolved.asset
	if asset == nil {
		return
	}
	previewURL := firstNonBlank(resolved.photo.URLs.Small, resolved.photo.URLs.Thumb)
	if target.visual != nil {
		target.visual.AssetQuery = target.query
		target.visual.AssetID = asset.PhotoID
		target.visual.LocalPath = asset.LocalPath
		target.visual.ImageURL = asset.ImageURL
		target.visual.PreviewURL = previewURL
		target.visual.SourceURL = asset.SourceURL
		target.visual.Attribution = asset.Attribution
	}
	if target.component != nil {
		target.component.AssetQuery = target.query
		target.component.AssetID = asset.PhotoID
		target.component.LocalPath = asset.LocalPath
		target.component.ImageURL = asset.ImageURL
		target.component.PreviewURL = previewURL
		target.component.SourceURL = asset.SourceURL
		target.component.Attribution = asset.Attribution
	}
}

func resolvedExistingBackgroundAsset(workDir string, target plannedBackgroundTarget) (resolvedBackgroundAsset, bool) {
	localPath := ""
	assetID := ""
	imageURL := ""
	previewURL := ""
	sourceURL := ""
	attribution := ""
	if target.visual != nil {
		localPath = target.visual.LocalPath
		assetID = target.visual.AssetID
		imageURL = target.visual.ImageURL
		previewURL = target.visual.PreviewURL
		sourceURL = target.visual.SourceURL
		attribution = target.visual.Attribution
	}
	if target.component != nil {
		localPath = target.component.LocalPath
		assetID = target.component.AssetID
		imageURL = target.component.ImageURL
		previewURL = target.component.PreviewURL
		sourceURL = target.component.SourceURL
		attribution = target.component.Attribution
	}
	if !taskLocalAssetExists(workDir, localPath) {
		return resolvedBackgroundAsset{}, false
	}
	asset := &unsplash.DownloadedAsset{
		PhotoID:     assetID,
		LocalPath:   localPath,
		ImageURL:    imageURL,
		SourceURL:   sourceURL,
		Attribution: attribution,
	}
	photo := unsplash.Photo{URLs: unsplash.PhotoURLs{Small: previewURL, Thumb: previewURL}}
	return resolvedBackgroundAsset{photo: photo, asset: asset}, true
}

func applyResolvedImageAsset(target plannedImageAssetTarget, resolved resolvedBackgroundAsset) {
	asset := resolved.asset
	if asset == nil || target.component == nil {
		return
	}
	previewURL := firstNonBlank(resolved.photo.URLs.Small, resolved.photo.URLs.Thumb)
	target.component.AssetQuery = target.query
	target.component.AssetID = asset.PhotoID
	target.component.LocalPath = asset.LocalPath
	target.component.ImageURL = asset.ImageURL
	target.component.PreviewURL = previewURL
	target.component.SourceURL = asset.SourceURL
	target.component.Attribution = asset.Attribution
}

func isBackgroundPlan(assetPurpose, imagePosition, role string) bool {
	purpose := strings.ToLower(strings.TrimSpace(assetPurpose))
	position := strings.ToLower(strings.TrimSpace(imagePosition))
	role = strings.ToLower(strings.TrimSpace(role))
	return purpose == "background" || position == "background" || role == "background" || role == "hero_photo"
}

func taskLocalAssetExists(workDir, localPath string) bool {
	localPath = strings.TrimSpace(localPath)
	if localPath == "" {
		return false
	}
	path := localPath
	if !filepath.IsAbs(path) {
		path = filepath.Join(workDir, path)
	}
	path = filepath.Clean(path)
	relative, err := filepath.Rel(workDir, path)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular() && info.Size() > 0
}

func firstNonBlank(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}
