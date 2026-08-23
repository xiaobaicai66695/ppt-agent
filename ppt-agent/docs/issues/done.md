# Done

## 2026-08-23

- ID: 20260823-reflexion-plan-review-gate
- Type: feature/fix
- Scope: DeckSpec planning, manifest commit gate, image-background planning quality
- Completed: 2026-08-23 21:22 Asia/Shanghai
- Changes:
  - Added `review_tasks_manifest` as a deterministic Reflexion-style planner review tool for `tasks.draft.json`.
  - Required reviewed and current draft fingerprints before `update_tasks_manifest(mode="commit")` can publish `tasks.json`.
  - Updated Planner instructions to run `initialize/patch -> review -> refine -> review -> commit`, with up to 3 review rounds.
  - Added a planning quality gate that prefers every slide to have an executable background image plan, using downloaded `local_path` when image search is available.
- Verification:
  - Local: `go test ./pkg/agent/deck` passed.
  - Local: `go test ./pkg/agent/deck ./pkg/tools/image ./pkg/agent/utils` passed.
  - Local: `go build ./...` passed.
- Deployment:
  - Target: `remote-dev:/ppt/ppt-agent/backend`
  - Backend process: restarted from PID `1766222` to PID `2519154`, command `./ppt-agent-linux -mode web -addr :8080`.
  - Health: `GET http://127.0.0.1:8080/api/health` returned 200 with `{"status":"ok"}`.
  - Templates: `GET http://127.0.0.1:8080/api/templates` returned 200 and returned preset metadata.
  - Binary verification: deployed binary contains `review_tasks_manifest` and the commit guard string.
- Cleanup:
  - Removed local `ppt-agent-linux` build artifact.
  - Removed this deployment's `/tmp/ppt-agent-linux.20260823*` files from the server.

## 2026-08-21

- ID: 20260821-inline-chat-tool-previews
- Type: feature
- Scope: frontend conversation stream, inline tool preview cards
- Completed: 2026-08-21 19:03 Asia/Shanghai
- Changes:
  - Added `deriveInlineConversationItems` and `deriveInlineToolPreviews` to merge assistant/user messages with real `tool_start/tool_end` runtime events by timestamp.
  - Rendered tool calls as expandable AI-row cards inside `ConversationComposer`, including args/result previews, source links, and searched image thumbnails.
  - Kept the existing runtime diagnostics panel as full trace while making the main chat stream show the same tool activity inline.
- Verification:
  - Local: `npm run test -- workbench` passed with 18 tests.
  - Local: `npm run build` passed.
- Deployment:
  - Target: `remote-dev:/ppt/ppt-agent`
  - Frontend: `GET http://127.0.0.1:8080/` returned 200 and served new `dist` asset `/assets/index-Bbu_foc6.js`.
  - Health: `GET http://127.0.0.1:8080/api/health` returned 200 with `{"status":"ok"}`.
  - Backend restart: not required; frontend-only change.
- Cleanup:
  - Removed `/tmp/ppt-agent-frontend-inline-tools-20260821190108.tgz`.

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
