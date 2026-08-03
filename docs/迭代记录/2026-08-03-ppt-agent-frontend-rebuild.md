# 2026-08-03 PPT Agent 前端工作台重构

## 迭代目标

在基线提交 `7289c60` 之后全面重构 Vue 前端，保留 REST、SSE、认证、任务和模板契约，把多个风格分裂的页面统一成可持续维护的 PPT 生产工作台。

## 已完成

- 建立 OpenSpec change：`openspec/changes/ppt-agent-frontend-rebuild/`，对应 TODO `PPT-UI-001`。
- 参考 Gamma、Canva、Pitch、Linear、Notion 的成熟交互模式，形成 prompt-first 创建、稳定编辑器区域、交付优先 Dashboard 和渐进披露诊断信息的设计方向。
- 新增 `AppShell` 与全局 semantic tokens，桌面使用近黑窄导航，平板和手机使用抽屉。
- 重构 Home、Auth、Compose、Dashboard、Admin 五个主页面。
- 引入 Lucide Vue，移除本轮触及页面的手绘 SVG、装饰渐变、光斑和 emoji 图标。
- 重做 Sidebar、ProgressBar、SlidePreviewCard，明确任务、阶段、缩略图和下载状态。
- 删除 Dashboard/Admin 中被覆盖的历史 CSS，并将最终规则写入 `ppt-agent/frontend/design-system/ppt-agent/MASTER.md`。

## 关键决策

- 采用浅灰画布、白色工作面、近黑导航和青绿/蓝/珊瑚多语义色，不采用常见 AI 紫渐变或深色电影感。
- Home 首屏是可直接工作的创建入口，不再是营销 landing page。
- Dashboard 的阅读顺序固定为任务上下文、用户进度、已就绪页面、任务活动、运行诊断。
- 日志、RuntimeMeta 和 Timeline 保留，但默认次级折叠；手机聊天栏参与布局，不覆盖内容。
- 不引入 Tailwind 或完整组件框架，继续使用 Vue 3 scoped CSS 和现有 API 类型。

## 验证

- `npm run build` 通过。
- Playwright + Chrome 覆盖 375、768、1024、1440，5 个主页面共 20 组截图，横向溢出均为 0，页面运行时异常为 0。
- 通过本地 route mock 补查 Compose、Dashboard、Admin 完整数据态。
- 验证主导航和任务抽屉完整进出视口，验证手机聊天栏不遮挡幻灯片。
- `openspec validate ppt-agent-frontend-rebuild --strict` 作为最终规格校验。

## 关联内容

- `docs/issues/todo.md`
- `docs/research/2026-08-03-ppt-workbench-ui-rebuild.md`
- `openspec/changes/ppt-agent-frontend-rebuild/`
- `ppt-agent/frontend/design-system/ppt-agent/MASTER.md`
