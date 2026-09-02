# 独立 skill 验证说明

本文档用于验证 `ppt-deck-planner` 在脱离 `ppt-agent` 后端时是否仍能作为独立 skill 工作。

## 1. 语法检查

外部 Agent 使用图片 CLI 时需要 Node.js 22 或更高版本。

先安装 Python 依赖：

```bash
python -m pip install -r requirements.txt
```

若在 `ppt-agent` 仓库根目录运行：

```bash
python -m pip install -r skills/ppt-deck-planner/requirements.txt
```

在 skill 根目录或仓库根目录运行：

```bash
python -m compileall -q generators
```

若在 `ppt-agent` 仓库根目录运行：

```bash
python -m compileall -q skills/ppt-deck-planner/generators
```

## 2. 单测

当前单测重点覆盖 `render_task.py` 和组件式页面渲染：

```bash
python -m unittest discover -s tests -v
```

若在 `ppt-agent` 仓库根目录运行：

```bash
python -m unittest discover -s skills/ppt-deck-planner/tests -v
```

## 3. DeckSpec 预检

纯文字 deck 必须显式设置 `visual_policy.mode="none"`，随后可直接预检：

```bash
python generators/validate_deck.py --work-dir examples/minimal --skills-dir ..
```

若在 `ppt-agent` 仓库根目录运行：

```bash
python skills/ppt-deck-planner/generators/validate_deck.py --work-dir skills/ppt-deck-planner/examples/minimal --skills-dir skills
```

预检会检查：

- `content_type` 是否合法。
- 组件类型是否合法。
- 组件是否写入坐标、字号、颜色、边距、透明度等禁止字段。
- 显式图片路径是否存在。
- 是否出现旧 `asset:` id。
- KPI 和图表组件是否包含必要结构化字段。

## 4. 生成整套 PPTX 冒烟

```bash
python generators/render_deck.py --work-dir examples/minimal --skills-dir .. --output deck.pptx
```

若在 `ppt-agent` 仓库根目录运行：

```bash
python skills/ppt-deck-planner/generators/render_deck.py --work-dir skills/ppt-deck-planner/examples/minimal --skills-dir skills --output deck.pptx
```

## 5. 生成 1 页 PPTX 冒烟

准备一个临时工作目录，放入最小 `tasks.json`。独立调用方可以自行决定工作目录位置；下面仅是示例。

```json
{
  "title": "Skill 独立渲染冒烟",
  "tasks": [
    {
      "task_id": "slide_01",
      "page_index": 1,
      "title": "独立 skill 的最小闭环",
      "content_type": "content_slide",
      "content_plan": {
        "summary": "用一页内容验证 DeckSpec 可以被 render_task.py 消费。",
        "slide_intent": "说明 skill 在没有后端 workflow 时仍能渲染语义内容。",
        "components": [
          {
            "type": "headline",
            "text": "独立 skill 只消费结构化语义和显式本地图片路径"
          },
          {
            "type": "bullet_list",
            "items": [
              "Planner 负责把用户需求转成完整 DeckSpec 内容字段，不写运行态字段。",
              "调用方决定图片下载、保存和路径写入策略，skill 不绑定固定 asset 目录。",
              "没有有效本地图片时，生成器渲染文本、卡片、图表或浅色面板作为语义表达。"
            ]
          }
        ]
      }
    }
  ]
}
```

然后运行：

```bash
python generators/render_task.py --work-dir <work-dir> --skills-dir <skill-parent-dir> --task-id slide_01
```

其中：

- `<work-dir>` 是包含 `tasks.json` 的目录。
- `<skill-parent-dir>` 是包含 `ppt-deck-planner/` 的目录；例如在本仓库中是 `skills`。

预期结果：工作目录中生成该页 PPTX，且没有依赖 `assets/manifest.json`。若 `tasks.json` 显式写入图片路径或旧 `asset:` id，该路径必须真实可读，否则渲染应失败。

## 6. 外部 Agent 的 Unsplash 图片 CLI

该 CLI 只供 `ppt-agent` 项目外的 Agent 使用；项目内使用后端已有的图片搜索能力。先在 Unsplash Developers 控制台创建应用，在应用的 **Keys** 页面复制 **Access Key**。不要使用 Secret Key。

在 skill 根目录运行下列命令，随后按提示输入 Key；输入不会回显，也不会出现在命令参数中：

```bash
npm link
unsplash auth
unsplash fetch --work-dir <work-dir>
```

认证信息保存到 skill 根目录的 `auth.txt`，该文件已被 Git 忽略。需要清除认证时删除该文件。除显式 `visual_policy.mode="none"` 外，素材解析是固定步骤；在 `mode="required"` 下，每个非豁免页面都必须有物化后的背景 `visual_intent` 或前景 `image` 组件，且 `min_image_pages` 不能低于非豁免页数。纯文字例外必须显式写入 `visual_intent.role="clean_text_only"`、`search_status="skipped"` 和非空 `skip_reason`。背景 `visual_intent` 和前景 `image` 组件都必须声明查询、主体、构图和方向。`validate_deck.py` 与 `render_deck.py` 会拒绝 query-only、缺图和用过低覆盖数掩盖的素材计划。

## 7. 视觉验收

如果环境安装了 LibreOffice 和 Poppler，可以把 PPTX 转成 PDF/PNG 进行人工或自动检查：

```bash
soffice --headless --convert-to pdf --outdir <out-dir> <slide.pptx>
pdftoppm -png -r 144 <slide.pdf> <out-prefix>
```

检查重点：

- 页面能打开，页数正确。
- 标题、正文、来源栏不重叠。
- 没有 `[图片占位]`、`主题视觉` 或虚构素材文案。
- 有图片的页面应有真实本地路径与来源信息；无图片页面不应声明虚构图片路径。

## 8. 依赖说明

核心依赖：

- `python-pptx`
- `Pillow`

视觉验收依赖：

- LibreOffice
- Poppler
