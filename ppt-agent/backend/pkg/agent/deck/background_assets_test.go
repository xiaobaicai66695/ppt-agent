package deck

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"

	"github.com/cloudwego/ppt-agent/pkg/assets/unsplash"
)

type fakeBackgroundAssetClient struct {
	mu        sync.Mutex
	searches  []unsplash.SearchOptions
	downloads int
	searchErr map[string]error
}

func (f *fakeBackgroundAssetClient) Search(_ context.Context, options unsplash.SearchOptions) (*unsplash.SearchResponse, error) {
	f.mu.Lock()
	f.searches = append(f.searches, options)
	f.mu.Unlock()
	if f.searchErr != nil {
		if err := f.searchErr[options.Query]; err != nil {
			return nil, err
		}
	}
	photoID := fmt.Sprintf("photo-%d-%x", options.Page, len(options.Query))
	return &unsplash.SearchResponse{Results: []unsplash.Photo{{
		ID: photoID,
		URLs: unsplash.PhotoURLs{
			Regular: "https://images.unsplash.com/" + photoID + ".jpg",
			Small:   "https://images.unsplash.com/" + photoID + "-small.jpg",
		},
		Links: unsplash.PhotoLinks{HTML: "https://unsplash.com/photos/" + photoID},
		User:  unsplash.User{Name: "Test Photographer"},
	}}}, nil
}

func (f *fakeBackgroundAssetClient) Download(_ context.Context, photo unsplash.Photo, dir string) (*unsplash.DownloadedAsset, error) {
	f.mu.Lock()
	f.downloads++
	f.mu.Unlock()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	path := filepath.Join(dir, "unsplash_"+photo.ID+".jpg")
	if err := os.WriteFile(path, []byte("fake image"), 0o644); err != nil {
		return nil, err
	}
	return &unsplash.DownloadedAsset{
		PhotoID:     photo.ID,
		LocalPath:   path,
		ImageURL:    photo.URLs.Regular,
		SourceURL:   photo.Links.HTML,
		Attribution: "Photo by Test Photographer on Unsplash",
	}, nil
}

func TestMaterializePlannedBackgroundsReusesOneBackgroundPerContentType(t *testing.T) {
	workDir := t.TempDir()
	manifest := &TasksManifest{Tasks: []*TaskItem{
		{
			TaskID:      "page-1",
			ContentType: "content_slide",
			ContentPlan: &ContentPlan{VisualIntent: &VisualIntent{
				AssetPurpose: "background", AssetSubject: "diplomacy", AssetQuery: "light global diplomacy meeting wide landscape clean negative space", Role: "hero_photo",
			}},
		},
		{
			TaskID:      "page-2",
			ContentType: "content_slide",
			ContentPlan: &ContentPlan{VisualIntent: &VisualIntent{
				AssetPurpose: "background", AssetQuery: "enterprise automation dashboard wide landscape clean negative space", ImagePosition: "background",
			}},
		},
		{
			TaskID:      "page-3",
			ContentType: "image_text",
			ContentPlan: &ContentPlan{VisualIntent: &VisualIntent{
				AssetPurpose: "background", AssetQuery: "light energy infrastructure wide landscape clean negative space",
			}},
		},
	}}
	targets, err := collectPendingBackgroundTargets(workDir, manifest)
	if err != nil {
		t.Fatal(err)
	}
	client := &fakeBackgroundAssetClient{}
	count, err := materializePlannedBackgroundsWithClient(context.Background(), workDir, targets, client)
	if err != nil {
		t.Fatal(err)
	}
	if count != 3 || len(client.searches) != 2 || client.downloads != 2 {
		t.Fatalf("unexpected materialization count=%d searches=%d downloads=%d", count, len(client.searches), client.downloads)
	}
	first := manifest.Tasks[0].ContentPlan.VisualIntent
	second := manifest.Tasks[1].ContentPlan.VisualIntent
	third := manifest.Tasks[2].ContentPlan.VisualIntent
	if first.LocalPath == "" || second.LocalPath == "" || third.LocalPath == "" {
		t.Fatalf("expected downloaded paths, got %q %q %q", first.LocalPath, second.LocalPath, third.LocalPath)
	}
	if first.LocalPath != second.LocalPath {
		t.Fatalf("expected same content_type to reuse one background, got first=%q second=%q", first.LocalPath, second.LocalPath)
	}
	if first.LocalPath == third.LocalPath {
		t.Fatalf("expected different content_type to allow another background, got %q and %q", first.LocalPath, third.LocalPath)
	}
	gotQueries := map[string]bool{}
	for _, search := range client.searches {
		if search.Orientation != "landscape" || search.PerPage != 1 {
			t.Fatalf("unexpected search options: %#v", search)
		}
		if search.Page != 1 {
			t.Fatalf("content-type background queries should start from page 1, got %#v", search)
		}
		gotQueries[search.Query] = true
	}
	if !gotQueries["diplomacy"] || !gotQueries["energy"] {
		t.Fatalf("expected compact single-keyword searches, got %#v", client.searches)
	}
	if first.PreviewURL == "" || first.SourceURL == "" || first.Attribution == "" || first.Provider != "unsplash" || first.SearchStatus != "resolved" {
		t.Fatalf("expected traceable image metadata: %#v", first)
	}
	if !taskLocalAssetExists(workDir, first.LocalPath) {
		t.Fatalf("downloaded task-local background missing: %s", first.LocalPath)
	}
}

func TestMaterializePlannedBackgroundsUsesOneAssetForRepeatedContentType(t *testing.T) {
	workDir := t.TempDir()
	query := "light AI cloud architecture wide landscape clean negative space"
	manifest := &TasksManifest{Tasks: []*TaskItem{
		{TaskID: "page-1", ContentType: "kpi_dashboard", ContentPlan: &ContentPlan{VisualIntent: &VisualIntent{AssetPurpose: "background", AssetQuery: query}}},
		{TaskID: "page-2", ContentType: "kpi_dashboard", ContentPlan: &ContentPlan{VisualIntent: &VisualIntent{AssetPurpose: "background", AssetQuery: query}}},
	}}
	targets, err := collectPendingBackgroundTargets(workDir, manifest)
	if err != nil {
		t.Fatal(err)
	}
	client := &fakeBackgroundAssetClient{}
	count, err := materializePlannedBackgroundsWithClient(context.Background(), workDir, targets, client)
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 || len(client.searches) != 1 || client.downloads != 1 {
		t.Fatalf("unexpected materialization count=%d searches=%d downloads=%d", count, len(client.searches), client.downloads)
	}
	if client.searches[0].Query != "ai" {
		t.Fatalf("expected compact single-keyword background query, got %#v", client.searches[0])
	}
	first := manifest.Tasks[0].ContentPlan.VisualIntent.LocalPath
	second := manifest.Tasks[1].ContentPlan.VisualIntent.LocalPath
	if first == "" || second == "" || first != second {
		t.Fatalf("expected one shared background image, got %q and %q", first, second)
	}
}

func TestMaterializePlannedDeckAssetsDownloadsForegroundImageComponents(t *testing.T) {
	workDir := t.TempDir()
	manifest := &TasksManifest{Tasks: []*TaskItem{{
		TaskID: "page-1",
		ContentPlan: &ContentPlan{Components: []PlanComponent{
			{
				ID:           "scene-1",
				Type:         "image",
				AssetPurpose: "scene",
				AssetQuery:   "AI control room operators reviewing dashboard",
			},
		}},
	}}}
	backgroundTargets, err := collectPendingBackgroundTargets(workDir, manifest)
	if err != nil {
		t.Fatal(err)
	}
	imageTargets, err := collectPendingImageAssetTargets(workDir, manifest)
	if err != nil {
		t.Fatal(err)
	}
	client := &fakeBackgroundAssetClient{}
	backgroundCount, err := materializePlannedBackgroundsWithClient(context.Background(), workDir, backgroundTargets, client)
	if err != nil {
		t.Fatal(err)
	}
	imageCount, err := materializePlannedImageAssetsWithClient(context.Background(), workDir, imageTargets, client)
	if err != nil {
		t.Fatal(err)
	}
	if backgroundCount != 0 || imageCount != 1 {
		t.Fatalf("unexpected counts background=%d image=%d", backgroundCount, imageCount)
	}
	component := manifest.Tasks[0].ContentPlan.Components[0]
	if component.LocalPath == "" || component.SourceURL == "" || component.Attribution == "" || component.Provider != "unsplash" || component.SearchStatus != "resolved" {
		t.Fatalf("expected foreground image metadata: %#v", component)
	}
	if !taskLocalAssetExists(workDir, component.LocalPath) {
		t.Fatalf("downloaded task-local image missing: %s", component.LocalPath)
	}
}

func TestMaterializePlannedDeckAssetsCollapsesExistingBackgroundsWithoutClient(t *testing.T) {
	workDir := t.TempDir()
	firstPath := writeFakeAsset(t, workDir, "first.jpg")
	secondPath := writeFakeAsset(t, workDir, "second.jpg")
	thirdPath := writeFakeAsset(t, workDir, "third.jpg")
	manifest := &TasksManifest{Tasks: []*TaskItem{
		{TaskID: "page-1", ContentType: "content_slide", ContentPlan: &ContentPlan{VisualIntent: &VisualIntent{AssetPurpose: "background", AssetQuery: "light AI office wide landscape", LocalPath: firstPath, Provider: "unsplash", SearchStatus: "resolved"}}},
		{TaskID: "page-2", ContentType: "image_text", ContentPlan: &ContentPlan{VisualIntent: &VisualIntent{AssetPurpose: "background", AssetQuery: "light AI city wide landscape", LocalPath: secondPath, Provider: "unsplash", SearchStatus: "resolved"}}},
		{TaskID: "page-3", ContentType: "content_slide", ContentPlan: &ContentPlan{VisualIntent: &VisualIntent{AssetPurpose: "background", AssetQuery: "light robot factory wide landscape", LocalPath: thirdPath, Provider: "unsplash", SearchStatus: "resolved"}}},
	}}
	counts, err := MaterializePlannedDeckAssets(context.Background(), workDir, manifest)
	if err != nil {
		t.Fatal(err)
	}
	if counts.Backgrounds != 3 || counts.Images != 0 {
		t.Fatalf("unexpected counts: %#v", counts)
	}
	if manifest.Tasks[2].ContentPlan.VisualIntent.LocalPath != firstPath {
		t.Fatalf("same content_type should reuse first background, got %q", manifest.Tasks[2].ContentPlan.VisualIntent.LocalPath)
	}
}

func TestMissingMaterializedBackgroundPagesRejectsQueryOnlyPlan(t *testing.T) {
	workDir := t.TempDir()
	manifest := &TasksManifest{Tasks: []*TaskItem{
		{TaskID: "page-1", PageIndex: 1, ContentPlan: &ContentPlan{VisualIntent: &VisualIntent{AssetPurpose: "background", AssetQuery: "industry"}}},
		{TaskID: "page-2", PageIndex: 2, ContentPlan: &ContentPlan{VisualIntent: &VisualIntent{Role: "clean_text_only"}}},
	}}
	if got := missingMaterializedBackgroundPages(workDir, manifest); !reflect.DeepEqual(got, []int{1}) {
		t.Fatalf("missing background pages = %#v, want [1]", got)
	}
}

func TestMaterializePlannedBackgroundsReturnsLLMRevisionRequestForHTTP410(t *testing.T) {
	workDir := t.TempDir()
	manifest := &TasksManifest{Tasks: []*TaskItem{
		{TaskID: "page-1", PageIndex: 1, ContentType: "content_slide", ContentPlan: &ContentPlan{VisualIntent: &VisualIntent{AssetPurpose: "background", AssetQuery: "missing conference hall"}}},
	}}
	targets, err := collectPendingBackgroundTargets(workDir, manifest)
	if err != nil {
		t.Fatal(err)
	}
	client := &fakeBackgroundAssetClient{searchErr: map[string]error{
		"missing": fmt.Errorf("unsplash API HTTP 410: Content removed"),
	}}
	_, err = materializePlannedBackgroundsWithClient(context.Background(), workDir, targets, client)
	var revision *assetQueryRevisionError
	if !errors.As(err, &revision) {
		t.Fatalf("expected assetQueryRevisionError, got %v", err)
	}
	if len(revision.Requests) != 1 {
		t.Fatalf("revision requests = %#v", revision.Requests)
	}
	request := revision.Requests[0]
	if !reflect.DeepEqual(request.TaskIDs, []string{"page-1"}) || !reflect.DeepEqual(request.PageIndexes, []int{1}) || !reflect.DeepEqual(request.Queries, []string{"missing"}) {
		t.Fatalf("unexpected revision request: %#v", request)
	}
}

func TestMaterializePlannedImageAssetsSkipsRecoverableSearchFailures(t *testing.T) {
	workDir := t.TempDir()
	query := "removed press briefing photo"
	manifest := &TasksManifest{Tasks: []*TaskItem{{
		TaskID: "page-1",
		ContentPlan: &ContentPlan{Components: []PlanComponent{{
			ID:           "scene-1",
			Type:         "image",
			AssetPurpose: "scene",
			AssetQuery:   query,
		}}},
	}}}
	targets, err := collectPendingImageAssetTargets(workDir, manifest)
	if err != nil {
		t.Fatal(err)
	}
	client := &fakeBackgroundAssetClient{searchErr: map[string]error{
		query: fmt.Errorf("unsplash API HTTP 410: Content removed"),
	}}
	count, err := materializePlannedImageAssetsWithClient(context.Background(), workDir, targets, client)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("expected no foreground images to be materialized, got %d", count)
	}
	component := manifest.Tasks[0].ContentPlan.Components[0]
	if component.LocalPath != "" {
		t.Fatalf("recoverable failure should leave failed image unset, got %q", component.LocalPath)
	}
	if component.AssetQuery != query {
		t.Fatalf("expected original image query to be preserved, got %q", component.AssetQuery)
	}
}

func TestCollectPendingBackgroundTargetsSkipsExistingAndCleanTextPlans(t *testing.T) {
	workDir := t.TempDir()
	existing := filepath.Join(workDir, "assets", "images", "existing.jpg")
	if err := os.MkdirAll(filepath.Dir(existing), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(existing, []byte("image"), 0o644); err != nil {
		t.Fatal(err)
	}
	manifest := &TasksManifest{Tasks: []*TaskItem{
		{TaskID: "existing", ContentPlan: &ContentPlan{VisualIntent: &VisualIntent{AssetPurpose: "background", AssetQuery: "query", LocalPath: existing}}},
		{TaskID: "clean", ContentPlan: &ContentPlan{VisualIntent: &VisualIntent{Role: "clean_text_only"}}},
	}}
	targets, err := collectPendingBackgroundTargets(workDir, manifest)
	if err != nil {
		t.Fatal(err)
	}
	if len(targets) != 0 {
		t.Fatalf("expected no pending backgrounds, got %#v", targets)
	}
}

func writeFakeAsset(t *testing.T, workDir, name string) string {
	t.Helper()
	path := filepath.Join(workDir, "assets", "images", name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("image"), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}
