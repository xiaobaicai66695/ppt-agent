# Done

## 2026-08-21

- ID: 20260821-image-tool-preview-and-embed
- Type: feature/fix
- Scope: image search runtime preview, component image asset contract, PPT image embedding, long narrative capacity
- Completed: 2026-08-21 18:05 Asia/Shanghai
- Changes:
  - Added `local_path`, `image_url`, `preview_url`, `source_url`, and `attribution` to `content_plan.visual_intent` and image components.
  - Extracted real `search_images` tool results into runtime metadata `image_results` for frontend expandable previews.
  - Rendered local image assets from task work dir into `image_text` and `image_hero` PPT slides instead of placeholder text.
  - Expanded `argument_block` and image-text narrative target length in planner/skill/component contracts.
- Verification:
  - Local: `go test ./pkg/agent/deck ./pkg/agent/utils ./pkg/tools/image` passed.
  - Local: bundled Python `-m unittest skills.visual_designer.tests.test_render_task_components` passed.
  - Local: `npm run build` passed.
  - Local: `go build ./...` passed.
  - Local full `go test ./...` blocked only by missing local `pdftoppm` in `pkg/tools/qa`.
- Deployment:
  - Target: `remote-dev:/ppt/ppt-agent`
  - Backend process: PID `1766222`, command `./ppt-agent-linux -mode web -addr :8080`
  - Health: `GET http://127.0.0.1:8080/api/health` returned 200 with `{"status":"ok"}`
  - Templates: `GET http://127.0.0.1:8080/api/templates` returned 200.
  - Frontend: `GET http://127.0.0.1:8080/` returned 200 and served new `dist` assets.
  - Online render smoke: temporary `image_text` task with `assets/images/scene.png` generated PPT successfully; package contained `ppt/media/*` and caption text.
- Cleanup:
  - Removed this deployment's `/tmp/ppt-agent-linux.20260821175912`, `/tmp/ppt-agent-deploy-20260821175912.tgz`, and temporary render smoke workspace.
