# 外部 Agent 的 Unsplash 图片 CLI

本说明只适用于脱离 `ppt-agent` 项目的独立 Agent。`ppt-agent` 项目内的 Agent 必须使用后端已有的图片搜索与下载链路，不能读取本 skill 的 `auth.txt`，也不能调用本 CLI。

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

成功注册命令后，认证时会明确显示以下提示。粘贴或输入 Access Key 后直接按 Enter；看不到输入字符是正常的保护行为：

```text
accessToken（输入后不会回显）:
```

如果 PowerShell 提示找不到 `unsplash`，说明 `npm link` 没有在 skill 根目录成功完成。回到 skill 根目录重新运行 `npm link`，再用 `Get-Command unsplash` 确认命令已注册。

要清除认证，只需删除 skill 根目录的 `auth.txt`。

## 下载已规划的图片

先查看顶层 `visual_policy`：`mode="required"` 时，所有非 `clean_text_only` 页面都必须拥有可下载背景或前景视觉计划；`mode="none"` 才可完全跳过认证和下载。`fetch` 同时扫描 `content_plan.visual_intent` 和 `components[].type="image"`，不会只下载背景。

需要图片时，先在 `visual_intent` 或 `image` 组件中填写 `asset_purpose`、`asset_query`、`asset_subject`、`composition` 与 `orientation`，再运行：

```bash
unsplash fetch --work-dir <work-dir>
```

`unsplash` 命令只面向项目外的独立 Agent；`fetch` 会把本地路径、来源链接和署名写回全部已声明的视觉资产。下载失败时不会写回部分 manifest；修正查询后重试即可。
