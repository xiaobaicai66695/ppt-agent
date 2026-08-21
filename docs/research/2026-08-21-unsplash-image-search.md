# Unsplash 图片搜索接入预研与验证

## 背景

Pixabay 注册和 API 权限验证受阻，改用 Unsplash 作为第一版专题图片搜索 provider。目标是把图片搜索放在 Planner 可调用的工具层，避免把第三方 API 细节写进 prompt 或 Python 生成器。

## 官方接口结论

- 搜索接口为官方公开 API：`GET https://api.unsplash.com/search/photos`。
- 公开搜索使用 `Authorization: Client-ID <Access Key>`；不需要用户 OAuth 登录，也不需要 Secret key。
- `401 Unauthorized` 的官方响应 `OAuth error: The access token is invalid` 表示 Access Key 未被接受，不表示访问了错误的文档接口。
- `download_location` 用于记录下载事件；下载到本地后仍保留 Unsplash 来源页和摄影师署名。
- 参考：[Unsplash API Documentation](https://unsplash.com/documentation)。

## 真实验证

验证时间：2026-08-21。

1. `remote-dev` 网络探测：
   - `api.unsplash.com` 可建立连接并返回 HTTP。
   - `images.unsplash.com` 返回 HTTP 200。
2. 首次 key：
   - 标准 Authorization 请求返回 401。
   - 使用 `client_id` 查询参数仍返回 401。
   - 结论：不是接口路径或服务器网络问题，而是 key 未开放/无效。
3. 用户重新开放 API 权限后：
   - `GET /search/photos` 返回 200。
   - `query=drone` 返回 `total=10000`，Go 客户端解析 3 条结果。
   - Eino `search_images` Tool 返回结构化 provider、图片、来源和署名字段。
   - `download=true` 成功触发下载追踪并将 `regular` 图片写入本地任务目录。
4. 2026-08-21 部署复测：
   - `remote-dev:/ppt/ppt-agent/backend` 新进程 PID `1675918`，`/api/health` 返回 200。
   - Client 使用背景代理词搜索返回 3 条结果；Tool 使用实景代理词返回 3 条结果并成功下载 3 张图片。
   - 探针二进制和临时素材目录已清理。

## 实现边界

```text
Planner
  └── search_images (Eino Tool)
        └── pkg/assets/unsplash (HTTP client)
              ├── Search /search/photos
              ├── track download_location
              └── download regular image + attribution
```

- 无多模态模型时，Planner 先把页面视觉角色转换为可见主体或语义代理，再生成 provider-ready 的 `asset_query`；用户标题本身不作为默认搜索词。
- `background` 专门描述整页氛围、宽幅、低细节和文字留白；`scene` 描述 PPT 中插入的具体对象、动作和环境；`evidence` 只承担视觉佐证，事实仍由 `search` 工具提供。
- 当主题和可检索图片没有精确对应关系时，使用相邻的环境、物体或动作作为视觉代理，并在 `asset_subject` 记录转换结果，在 `composition` 记录留白方向和主体位置。
- `search_images` 只在配置 `UNSPLASH_ACCESS_KEY` 时注册到 Planner；未配置时仍保留图片规划字段，但不会诱导模型调用未注册工具。
- `download=true` 将图片写入当前任务目录的 `assets/images`，同时返回来源页、摄影师和署名，供后续素材登记或渲染接入使用。

- `UNSPLASH_ACCESS_KEY` 只从环境变量读取。
- 未配置 key 时不向 Planner 注册 `search_images`，prompt 只保留 `asset_query` 规划字段。
- Tool 默认只搜索；需要生成器使用本地图片时显式传 `download=true`。
- 下载目录限定在当前任务工作目录内，防止工具参数写出任务目录。
- 结果不返回 Access Key 或 Secret key。

## 代码与验证入口

- 客户端：`ppt-agent/backend/pkg/assets/unsplash`
- Tool：`ppt-agent/backend/pkg/tools/image`
- Planner 注册：`ppt-agent/backend/pkg/agent/deck/agent.go`
- Go 客户端探针：`go run ./cmd/unsplash_probe -query drone`
- Tool 探针：`go run ./cmd/unsplash_tool_probe -query drone`
- 单测：`go test ./pkg/assets/unsplash ./pkg/tools/image ./pkg/agent/deck ./pkg/prompts`
- 规划契约回归：`TestPlanComponentUnmarshalKeepsImagePlanningContract`、`TestMainAgentPromptAdvertisesImageSearchOnlyWhenConfigured`

## 风险与后续

- Access Key 已在聊天截图中出现，建议在 Unsplash 控制台重新生成并只写入服务器 `.env`，不要提交仓库。
- Unsplash 结果是视觉素材候选，不承担事实准确性；事实数据继续使用 `search` 工具。
- 当前完成搜索和单图下载；后续可补充缓存、去重、素材选择评分、任务级素材 manifest，以及 PPT 备注/元数据中的统一署名。
