## Context

前端是 Vue 3 + TypeScript + Vite，路由页面按需加载，但页面各自维护大量 scoped CSS，Home、Auth、Compose、Dashboard 和 Admin 缺少共享应用壳。Compose 与 Dashboard 已包含完整业务状态、API 和 SSE 逻辑，本次应保留这些契约并重构展示层。视觉参考来自成熟 AI 创作、演示文稿编辑器和生产力工作台，但不能复制外部品牌、素材或受保护界面。

## Goals / Non-Goals

**Goals:**

- 用一致的导航、顶栏、token、图标和交互状态连接所有业务页面。
- 让 Home 成为可直接开始工作的创建入口，让 Compose 和 Dashboard 分别成为编辑器与交付工作台。
- 用真实模板和幻灯片预览承担主要视觉，减少装饰、同质卡片和说明性文案。
- 在 375、768、1024、1440 视口保持完整业务能力、清晰层级和无横向溢出。
- 降低后续维护成本，使页面布局与全局视觉规则分离。

**Non-Goals:**

- 不改变后端 API、SSE 事件、任务状态和模板契约。
- 不实现浏览器内 PPTX 自由编辑器或拖拽画布引擎。
- 不引入 Tailwind、shadcn 或完整组件框架。
- 不复制 Gamma、Canva、Pitch、Linear 或 Notion 的品牌资产和像素级界面。

## Decisions

### Decision 1: Shared application shell with route-aware navigation

新增 `AppShell` 和共享导航配置，Home、Compose、Dashboard、Admin 使用相同窄侧栏与上下文顶栏。桌面侧栏常驻，平板和手机使用抽屉/紧凑顶栏。Auth 保持独立壳层，但复用同一 token 和品牌标识。

选择共享壳而不是继续在每页复制导航，是因为位置稳定比单页视觉更能提升产品感，也能统一移动端、焦点和安全区处理。

### Decision 2: Neutral canvas with multi-semantic accents

设计系统使用浅灰画布、白色工作面和近黑导航；青绿用于主要动作，蓝色用于链接/信息，珊瑚色用于提醒，避免紫色渐变和暗蓝单色主题。圆角限制在 4/6/8px，阴影只用于抽屉、菜单和可交互预览。

本地 UI 数据库建议 AI 紫与深色电影感，但该方案会强化用户已指出的 AI 模板感，也不适合高频生产工具，因此仅采用其密度、可访问性和动效规则，不采用推荐配色与装饰效果。

### Decision 3: Prompt-first home and editor-style Compose

Home 首屏提供需求输入、创建入口、模板预览和最近工作路径，不再展示长营销 hero。Compose 保留现有数据与编辑动作，桌面重排为资源栏、页面轨道、属性区，主要生成动作位于稳定顶栏；窄屏属性区使用全屏 sheet。

### Decision 4: Delivery-first Dashboard hierarchy

Dashboard 顺序固定为任务标题/主操作、用户进度、就绪幻灯片、活动与继续对话。RuntimeMeta 与 Timeline 留在折叠诊断区。任务导航在桌面为窄列表，窄屏为抽屉；聊天在手机使用底部工作区但不得遮挡内容。

### Decision 5: Lucide icons and semantic primitives

新增 `lucide-vue-next`，本轮触及页面使用同一 1.75-2px outline 图标。按钮、导航、dialog、details、input 保持原生语义，图标按钮具有 `aria-label`/`title`，最小交互区域 44px。

### Decision 6: Preserve behavior, split visual responsibilities

Compose/Dashboard 的 API、SSE、computed 与事件处理函数尽量不动；先替换全局 token 和模板结构，再拆分无状态展示组件。这样可在大幅视觉变化时限制业务回归面。

## Risks / Trade-offs

- [Risk] 大型 SFC 模板和 CSS 重构容易误伤已有事件绑定。 -> 保留 script 区，分页面构建并在每步执行 `npm run build` 与浏览器交互检查。
- [Risk] 全局 token 改名会破坏旧 scoped CSS。 -> 提供短期兼容别名，页面迁移完成后再删无引用 token。
- [Risk] 首页真实任务/模板数据依赖登录和后端。 -> 使用现有 API 的真实加载、空态和错误态，不伪造生产数据。
- [Risk] 新依赖增加 bundle。 -> 仅使用可 tree-shake 的 Lucide Vue 图标，不引入完整 UI 框架。
- [Risk] 三栏编辑器在小屏空间不足。 -> 375/768 使用单主区 + sheet，不压缩成不可操作的三栏。

## Migration Plan

1. 建立 token、Lucide 依赖和共享 AppShell，保留旧 token 兼容层。
2. 重构 Home/Auth，确认新视觉系统和路由导航。
3. 重构 Compose、Dashboard 及其共享组件，保持业务脚本。
4. 重构 Admin，清理触及范围的手写 SVG 和旧装饰样式。
5. 构建并在 375/768/1024/1440 进行截图、交互、溢出和可访问性检查。
6. 若发生严重回归，可回退到基线提交 `7289c60`；无后端数据迁移。

## Open Questions

- 暗色模式暂不作为本轮交付要求；先完成高质量浅色工作台，后续基于同一 semantic token 单独规划。
