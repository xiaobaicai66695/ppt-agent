# Unsplash 图片 CLI

本 CLI 是 `ppt-deck-planner` skill 的**通用 Agent**图片搜索与下载入口。`ppt-agent` 项目内的闲聊 Agent 改用 `search_images` Tool 展示图片候选；PPTPlanner 不直接调用 CLI 或图片 Tool，只输出视觉意图，再由审查后的确定性素材物化层执行搜索、下载与路径回填。

## 图片工具优先级

当 DeckSpec 使用默认的 `visual_policy.mode="required"` 时，这个 CLI 是 Agent 的**默认图片搜索与下载入口**：先规划每页 `visual_intent`，再用 `unsplash fetch` 物化图片。不要因为宿主提供了 `imagegen`、绘图或截图工具就跳过搜索并自行生成背景图。

图像生成仅在用户明确要求 AI 生成插画、概念图或艺术图时使用，且该图不能替代每页 required 背景。CLI 未注册、认证失败、网络失败或没有结果时，保留失败原因并修复后重试；禁止以生成图、虚构本地路径或伪造 Unsplash 来源作为回退。仅当另一个宿主工具实际搜索并下载 Unsplash 图片，且能回填同等的路径、来源和署名时，才可替代本 CLI。

## 获取 Access Key

1. 打开 Unsplash Developers 控制台并登录账号。
2. 创建或选择一个 application。
3. 在该 application 的 **Keys** 页面复制 **Access Key**。

不要使用或保存 **Secret Key**。Access Key 也不要写进 `tasks.json`、prompt、命令参数、日志或仓库文件。

## 配置认证

在 skill 根目录的交互式终端中，首次运行一次 `npm link` 注册命令，然后执行：

```bash
npm link
unsplash auth
```

按 `accessToken:` 提示输入 Access Key。输入内容不会回显，命令会将其保存为 skill 根目录的 `auth.txt`；该文件已在 `.gitignore` 中忽略。Windows 无需设置环境变量。

服务器部署中，先加载服务使用的 `.env`，再通过环境变量完成无交互认证；优先读取 `UNSPLASH_ACCESS_KEY`，兼容 `UNSPLASH_ACCESS_TOKEN`：

```bash
set -a
. /ppt/ppt-agent/backend/.env
set +a
node /ppt/ppt-agent/skills/ppt-deck-planner/scripts/unsplash.mjs auth --from-env
```

此命令不会在终端输出 Access Key；后端进程仍必须在其启动环境中保留同一个 `UNSPLASH_ACCESS_KEY`，供确定性素材物化层访问 Unsplash API。

成功注册命令后，认证时会明确显示以下提示。粘贴或输入 Access Key 后直接按 Enter；看不到输入字符是正常的保护行为：

```text
accessToken（输入后不会回显）:
```

如果 PowerShell 提示找不到 `unsplash`，说明 `npm link` 没有在 skill 根目录成功完成。回到 skill 根目录重新运行 `npm link`，再用 `Get-Command unsplash` 确认命令已注册。

要清除认证，只需删除 skill 根目录的 `auth.txt`。

## 下载已规划的图片

先查看顶层 `visual_policy`：默认必须为 `mode="required"`，每一页都必须拥有可下载的背景 `visual_intent`；`mode="none"` 只有在 `user_declined_background:true` 和 `decline_reason` 记录了用户明确拒绝背景图片时才可完全跳过认证和下载。**同一 `content_type` 的背景必须先规划同一个 `asset_query`，最终必须复用同一个本地图片路径。** `fetch` 会为每种页面类型只解析一张背景，并把路径、来源和署名回写到该类型的全部页面；它不会按页轮换背景。前景 `components[].type="image"` 仍按组件自身语义独立处理。

需要图片时，先在 `visual_intent` 或 `image` 组件中填写 `asset_purpose`、`asset_query`、`asset_subject`、`composition` 与 `orientation`，再运行：

```bash
unsplash fetch --work-dir <work-dir>
```

`fetch` 会把本地路径、来源链接和署名写回全部已声明的视觉资产。下载失败时不会写回部分 manifest；修正查询后重试即可。
