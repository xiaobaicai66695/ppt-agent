package deck

import "context"

// DownloadPPTAssets downloads images required by the selected PPT pages. It
// keeps the legacy MaterializePlannedDeckAssets implementation behind a plain
// business-action name.
func DownloadPPTAssets(ctx context.Context, workDir string, plan *TasksManifest) (MaterializedDeckAssetCounts, error) {
	return MaterializePlannedDeckAssets(ctx, workDir, plan)
}
