---
name: visual_designer
description: 为PPT幻灯片提供视觉设计指导。遵循本Skill生成设计JSON，调用Python工具生成PPT。
---

# Visual Designer

## 禁止

- 紫色渐变 + 白底（AI感配色）
- Inter/Roboto/Arial 英文字体
- 超过 3 种主色调、彩虹渐变
- 同一层级内容混用对齐方式
- 全屏背景图 + 半透明文字
- 页面无留白

## 颜色主题

```
商务正式: primary=#1A365D, accent=#63B3ED, bg=#FFFFFF, alt=#F7FAFC
科技技术: primary=#0D1B2A, accent=#00B4D8, bg=#FFFFFF, alt=#F0F9FF
创意艺术: primary=#6B46C1, accent=#D6BCFA, bg=#FAF5FF, alt=#F3E8FF
极简纯净: primary=#000000, accent=#6366F1, bg=#FFFFFF, alt=#FAFAFA
```

字体：中文=思源黑体/微软雅黑，英文=Segoe UI/Poppins。

## 布局路由

| 要点数量 | 布局 |
|---------|------|
| ≤ 3 条 | 居中大字号 + 图标 |
| 4-6 条 | 双栏分列 |
| ≥ 7 条 | 分组 + 小标题 |
| 含数据 | 左文右图/图表 |

| 内容类型 | 布局 |
|---------|------|
| 概念定义 | 左图右文 |
| 多维对比 | 双栏对比 |
| 时间序列 | 时间轴 |
| 流程步骤 | 分步图 |
| 金句引言 | 大字居中 |

---

## 文字溢出防护（几何体内嵌文字必须遵守）

1. **先估算**：中文文字宽度 ≈ `字数 × 字号 × 0.5`，英文 ≈ `字符数 × 字号 × 0.3`
2. **几何体宽度 ≥ 文字宽度 + 两倍边距**，宁可大不可小
3. **优先换行**，换行后仍超出才缩小字号（最小 14pt）
4. **绝对不允许文字超出几何体边界**

---

## 图片素材获取（必须执行）

**不要凭空生成图片**。每个内容页按以下步骤获取素材：

1. 提取 2-3 个核心关键词（如："AI芯片"、"数据分析"）
2. 调用图片搜索工具，搜索 "关键词 高清配图" 或 "关键词 site:unsplash.com"
3. 选择无版权、高质量、适合PPT的图片URL
4. 将 `image_url` 传给Python代码下载并插入

---

## 设计输出格式

```json
{
  "theme": "tech",
  "slides": [
    {
      "index": 1,
      "type": "title_slide",
      "content": { "title": "...", "subtitle": "..." },
      "style": { "primary_color": "#0D1B2A", "accent_color": "#00B4D8" },
      "visual_elements": ["..."],
      "image_url": ""
    },
    {
      "index": 2,
      "type": "content_slide",
      "content": { "title": "...", "bullets": [{ "text": "..." }] },
      "style": { "primary_color": "#0D1B2A", "accent_color": "#00B4D8", "line_height": "1.5x", "use_background_alt": false },
      "visual_elements": ["..."],
      "image_url": "https://..."
    }
  ]
}
```
