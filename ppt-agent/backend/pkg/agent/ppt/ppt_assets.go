package ppt

import "context"

// DownloadPPTAssets downloads images required by the selected PPT pages. It
// keeps the legacy MaterializePlannedPPTAssets implementation behind a plain
// business-action name.
func DownloadPPTAssets(ctx context.Context, workDir string, plan *TasksManifest) (MaterializedPPTAssetCounts, error) {
	return MaterializePlannedPPTAssets(ctx, workDir, plan)
}
