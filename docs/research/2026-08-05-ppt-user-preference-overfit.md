# PPT 用户偏好过拟合预研

## 背景

当前 PPT Agent 已有用户风格偏好学习能力：任务创建时会读取用户历史画像并注入 prompt，任务完成后会从生成结果中提取风格并回写画像。用户反馈的问题是：当新输入与历史常用场景差异较大时，系统仍会按历史场景填充 `tasks.json` / JSON 内容，表现为模板、配色、布局和内容叙事被历史偏好牵引。

这个问题不是单点 prompt 问题，更像“偏好注入没有相关性门控 + 偏好存储扁平化 + 更新策略强化历史高频模式”的组合效应。

## 目标与非目标

目标：

- 梳理用户偏好从采集、存储、推荐、注入到更新的现状链路。
- 判断导致跨场景过拟合的主要触发点。
- 给出分阶段技术方案，优先降低错误迁移，再考虑画像结构升级。
- 明确需要补充的验证样例。

非目标：

- 本预研不直接修改后端代码或 prompt 行为。
- 不重新设计完整推荐系统，也不引入新的向量数据库或外部服务。
- 不处理 PPT 视觉生成器本身的排版质量问题。

## 现状

### 数据流

```text
用户 query
  |
  v
web.handleCreateTask
  |-- styleStore.Get(uid).BuildStyleContext()
  |   直接把基础用户画像写入 cfg.StyleContext
  |
  v
task.CreateTask / deck.ProcessUserIntent
  |-- learning.Engine.ProcessTask(query, userID)
  |   |-- intent.Classifier 得到 Domain / SuggestedTemplates / SuggestedTheme / SuggestedPageCount
  |   |-- profileStore.GetEnhanced(userID)
  |   |-- router.ProfileMatcher.EnhanceWithProfile()
  |       把历史模板、主题、典型页数合入分类结果
  |
  v
deck.enhanceStyleContextWithProfile()
  再把增强画像写入 StyleContext
  |
  v
master_instruction.tmpl
  StyleContext 出现在主 prompt 第一行，影响模板选择与 tasks.json 规划
  |
  v
任务完成
  |-- web.updateUserStyleFromTask()
  |-- styleExtractor.ExtractFromPPTX() 或 ExtractFromTasks()
  |-- styleStore.UpdateWithTask()
  |-- learning.Updater / RecordSuccess 累计画像
```

### 关键实现点

- `ppt-agent/backend/pkg/web/handler.go:163` 在任务创建时直接注入 `UserProfile.BuildStyleContext()`。
- `ppt-agent/backend/pkg/style/profile.go:412` 的 `BuildStyleContext()` 会输出“请在生成PPT时遵循上述偏好”，对当前任务没有相关性判断。
- `ppt-agent/backend/pkg/agent/router/engine.go:198` 的 `EnhanceWithProfile()` 会把 `GetPreferredTemplates()` 前置到 `SuggestedTemplates`，并用历史典型页数影响当前页数。
- `ppt-agent/backend/pkg/agent/deck/agent.go:244` 会把推荐模板、推荐配色、推荐页数继续写入 `StyleContext`。
- `ppt-agent/backend/pkg/agent/deck/agent.go:304` 的 `enhanceStyleContextWithProfile()` 会再次注入历史配色、布局、模板风格、成功经验等高敏感字段。
- `ppt-agent/backend/pkg/style/profile.go:149` 的 `GetPreferredTemplates()` 会从全局成功模式和 `ContentTypes` 推断模板，未按当前领域过滤。
- `ppt-agent/backend/pkg/agent/learning/updater.go:230` 要求 LLM 对画像字段做追加式更新，`applyProfileUpdates()` 对 themes/colors/layout/content_types/domain_preferences 也是追加或累加。
- `ppt-agent/doc/learning-system-domain-match.md` 已经识别出“领域相关性判断”这个方向，但实现层仍未真正落地。

## 问题判断

### 1. 偏好字段没有区分“全局偏好”和“领域偏好”

语言风格、页数、详细程度这类习惯可以跨领域迁移；配色、模板、布局、背景、成功模式高度依赖 PPT 场景。当前画像是扁平结构，`PreferredThemes`、`PreferredColors`、`LayoutPreferences`、`SuccessPatterns` 被统一注入，导致“商业路演偏好”会污染“学术答辩/政务汇报/课程教学”等新任务。

### 2. 注入链路重复且指令权重偏高

同一类历史偏好至少有两次进入 prompt：

- web 层基础画像：`BuildStyleContext()`。
- deep intent 层增强画像：推荐信息 + `enhanceStyleContextWithProfile()`。

而主 prompt 模板把 `StyleContext` 放在第一行，早于“用户主题红线”和模板选择规则。模型在规划 `tasks.json` 时更容易把历史偏好当作高优先级上下文。

### 3. 路由推荐会提前改变候选模板顺序

`ProfileMatcher.EnhanceWithProfile()` 会把历史偏好模板 prepend 到 `SuggestedTemplates`。即使意图分类器从当前 query 识别出了新领域，历史模板也可能抢占前排，后续 `StyleContext` 又把这些模板写成“推荐模板”，强化模型选择。

### 4. 更新策略会强化高频历史而非校正场景切换

基础画像通过 `MergeWith()` 和 `appendUnique()` 保留历史高频项；LLM Updater 的 prompt 明确要求列表字段追加，计数字段累加。缺少“本次任务和历史偏好不相关时不要更新/不要迁移”的约束，也缺少 per-domain 画像桶，所以历史常用场景越多，惯性越强。

### 5. 设计文档与实现脱节

`ppt-agent/doc/learning-system-domain-match.md` 已提出字段敏感性分级、领域重叠度、只注入 global preferences 等方案。但当前代码中的 `EnhanceWithProfile()`、`BuildStyleContext()`、`enhanceStyleContextWithProfile()` 仍是全量/半全量注入。

## 市面成熟产品调研

调研时间：2026-08-05。调研对象以官方产品页、帮助中心和支持文档为主，重点看“成熟 PPT 生成工具如何处理品牌、用户偏好、上下文和跨场景迁移”。

### 调研对象

| 产品 | 偏好/品牌做法 | 对 PPT Agent 的启发 |
| --- | --- | --- |
| Microsoft Copilot for PowerPoint | 通过 PowerPoint 模板、组织品牌模板和 Brand Kit 约束生成。官方说明中，Copilot 主要使用 sample slides，其次使用模板结构和 layouts；Brand Kit 可补充 brand voice、tone、style、imagery。用户在 PowerPoint 中可以显式选择 Brand Kit 后再输入 prompt。参考：[Keep your presentation on-brand with Copilot](https://support.microsoft.com/en-us/powerpoint/copilot/keep-your-presentation-on-brand-with-copilot)、[Create and manage official Brand kits](https://support.microsoft.com/en-us/microsoft-365-copilot/create-and-manage-official-brand-kits-in-the-microsoft-365-copilot-app)、[FAQ about Copilot in PowerPoint](https://support.microsoft.com/en-us/powerpoint/frequently-asked-questions-about-copilot-in-powerpoint)。 | 偏好不是隐式全局记忆，而是“当前显式选择的模板/Brand Kit + 当前 prompt/文件”共同决定。品牌约束要有可选入口和可追溯来源。 |
| Canva | Brand Kit 集中管理 logos、colors、fonts、assets、Brand Templates 和 guidelines；Canva AI / Magic Design 可在设计生成后或生成过程中应用 Brand Kit。Canva 还支持从网站或 PDF 自动创建 Brand Kit。参考：[Set up Brand Kits](https://www.canva.com/help/brand-kit/)、[Generate on-brand designs with Brand Templates and Brand Kits](https://www.canva.com/help/create-on-brand-designs/)、[Set up your Brand Kit automatically](https://www.canva.com/help/brand-kit-builder/)。 | 品牌偏好是结构化资产库，不是自然语言长备注；多品牌/多项目时要隔离，不应把 A 场景品牌污染到 B 场景。 |
| Gamma | 用 custom theme 表达整体视觉风格，包含 colors、fonts、slide styles、accent images；workspace 内共享 custom themes，用户可在生成/编辑后切换或定制 theme。参考：[Can I add my own colors and fonts to Gamma?](https://help.gamma.app/en/articles/11029150-can-i-add-my-own-colors-and-fonts-to-gamma)、[How do I change my Gamma theme?](https://help.gamma.app/en/articles/10262646-how-do-i-change-my-gamma-theme)、[Teams and business options](https://help.gamma.app/en/articles/11594955-what-options-does-gamma-offer-for-teams-and-business)。 | 将“偏好”收敛为可命名、可选择、可共享的 theme，比把历史偏好散落在 prompt 中更稳定。 |
| Beautiful.ai | 强调 guided workflow、Smart Slides 和 Brand Control：先从 prompt 到 outline，再设计和 refine；Smart Slides 负责自动处理 spacing、alignment、hierarchy、chart layout；Teams/Enterprise 可定义 fonts、colors、logos、layouts，并锁定设计元素和管理权限。参考：[Beautiful.ai](https://www.beautiful.ai/)、[Customizable AI Presentation Slide Templates](https://www.beautiful.ai/slide-templates)、[AI Design Doesn't Replace Brand Control](https://www.beautiful.ai/blog/ai-design-doesnt-replace-brand-control-it-strengthens-it)。 | 成熟产品把“结构规划”和“视觉约束”拆开，并用版式系统兜住质量；偏好不直接决定内容 JSON，而是约束可编辑设计系统。 |
| Pitch | Pitch Agent 从 prompt、模板和附件生成 deck；官方介绍称其从真实模板生成，使用 custom layouts 反映品牌模式。新团队可输入网站域名，让 Agent 基于 website 生成 branded template，包括 colors、fonts、logo、image style；workspace library 管理 templates、videos、images、custom fonts。参考：[Pitch Agent](https://pitch.com/blog/introducing-pitch-agent)、[Create on-brand decks with Pitch's AI presentation maker](https://pitch.com/use-cases/ai-presentation-maker)、[Create a template](https://help.pitch.com/en/articles/3752837-create-a-template)、[Workspace Library](https://help.pitch.com/en/articles/6010928-organize-your-workspace-library)。 | 当前任务可用的 template/附件/域名比历史行为更重要；品牌可从当前项目来源抽取，并在进入生成前让用户确认。 |
| Presentations.AI | 官方称可从 topic、URL、document 生成 on-brand PPTX，也声明 AI 会学习用户的 structure、messaging、visual preferences，并从用户 refine/edit 中继续学习。它同时强调 website URL 可建立 rigid Brand Kit，管理员控制核心 themes。参考：[AI Presentation Maker](https://www.presentations.ai/ai-presentation-maker)、[Features and Capabilities](https://www.presentations.ai/features)、[Marketing solution](https://www.presentations.ai/solutions/marketing)。 | 少数产品会宣传“学习偏好”，但仍把 brand kit、URL、document、admin-controlled themes 作为硬约束。对 PPT Agent 来说，学习结果应低于当前输入和显式品牌资产。 |

### 共性模式

成熟产品的做法可以归纳为五类：

1. 显式品牌资产优先：通过 Brand Kit、Theme、Template、Workspace Library 管理 logo、font、color、layout、image style，而不是只靠历史自然语言偏好。
2. 当前任务上下文优先：prompt、上传文件、URL、参考文档、当前模板是生成的主要依据，历史偏好通常作为辅助。
3. 用户选择或确认：品牌包、模板、主题往往需要用户选择、管理员发布或从当前域名生成后再应用，降低跨场景误用。
4. 结构化约束而非 prompt 强命令：颜色、字体、版式、模板、权限、锁定元素进入设计系统或模板系统，而不是拼一句“遵循偏好”。
5. 生成后可迭代：多数工具强调 outline 预览、refine、chat edit、theme switch、Smart Slides 自动重排，避免一次生成后被历史偏好锁死。

### 对当前过拟合问题的启发

PPT Agent 目前更接近“隐式历史画像强注入”，而成熟产品更接近“显式当前品牌/模板弱到强约束”。因此治理方向应调整为：

- 把历史偏好降级为参考信号，永远低于当前 query、当前 outline、用户显式选择的 template/theme/background。
- 将高敏字段结构化：template、theme、colors、layout、brand assets 应进入可选择的 `BrandKit` / `ThemeProfile` / `DomainProfile`，不要作为全局自然语言塞进 `StyleContext`。
- 允许多品牌/多场景隔离：同一用户可以有 business、academic、government、education 等多个 profile bucket。
- 引入显式选择和可观测来源：StyleContext 中标明“来自当前选择”“来自同领域历史”“来自跨领域全局习惯”，并按来源赋权。
- 生成前或规划阶段优先让当前任务确定模板和配色；历史偏好只能在 domain 匹配且用户没有明确选择时补位。

## 方案选项

### 方案 A：短期门控，按当前领域过滤偏好注入

在不改数据库结构的前提下，先增加一个“偏好相关性过滤器”：

- 输入：当前 `ClassificationResult.Domain`、用户 query、`EnhancedProfile`。
- 输出：`ScopedPreferenceContext`，区分 `global` 与 `domain_specific`。
- 低敏字段总是可用：`LanguageTone`、`TypicalPageCount`、`ContentTone.DetailLevel`、`AnimationLevel` 可低权重注入。
- 高敏字段需匹配当前领域：`PreferredThemes`、`PreferredColors`、`LayoutPreferences`、`SuccessPatterns`、由 `ContentTypes` 推断的模板。
- 若当前领域为 `unknown` 或与历史 TopDomain 不匹配，只注入全局字段，不注入模板/配色/布局。
- 将 prompt 文案从“请遵循上述偏好”改为“以下为可参考偏好；若与当前主题冲突，以当前主题和用户显式要求为准”。

优点：改动小，能直接缓解过拟合；适合先做 direct 小步迭代。

不足：历史基础字段没有来源领域，过滤会偏保守；无法精细追踪不同领域内部的偏好。

### 方案 B：中期分层画像，建立 per-domain profile

把画像拆成：

- `global_preferences`：语言风格、详细程度、常用页数、动画强度。
- `domain_profiles[domain]`：模板、主题、配色、布局、背景、内容结构、成功模式。
- `recent_tasks`：保留最近 N 个任务的 domain、template、theme、page_count、quality_score，用于漂移判断和离线评估。

优点：语义正确，长期可维护；能支持同一用户跨多个 PPT 场景稳定工作。

不足：需要数据库兼容迁移、API 展示调整和一组回归测试；更适合进入 OpenSpec。

### 方案 C：关闭默认偏好注入，仅保留显式选择

短期可以在配置层关闭历史偏好进入主 Agent，只保留用户当前在前端选择的模板、主题、页数。

优点：风险最低，立刻避免历史污染。

不足：等于暂时牺牲学习系统价值；用户常用场景下的个性化收益消失。

## 推荐方案

推荐先走“方案 A + 少量可观测性”，后续再推进“方案 B”。

### 第一阶段：止血

- 在 `ProcessUserIntent()` 之后、写入 `StyleContext` 之前做偏好筛选。
- `ProfileMatcher.EnhanceWithProfile()` 不再无条件前置历史模板；只在当前 domain 与成功模式 domain 匹配时前置。
- `GetPreferredTheme()`/`GetPreferredTemplates()` 增加 domain-aware 变体，旧方法保留给兼容路径。
- web 层 `BuildStyleContext()` 不应绕过 intent domain 直接注入；至少应改成由 deep 层统一注入，避免重复。
- prompt 中明确当前 query 优先级高于历史偏好。

### 第二阶段：可验证

补一组后端单测和最小离线样例：

- 用户历史 5 次 `business`，当前 query 为 `academic`：不应推荐 `pitch-deck`、`charcoal_light` 作为首选。
- 用户历史 5 次 `business`，当前 query 仍为 `business`：可复用历史模板/配色。
- 用户历史页数为 16，当前 query 明确“做 6 页短分享”：页数应尊重当前 query。
- 用户历史有“深色背景” special note，当前 query 为“儿童课程活泼课件”：不应注入深色背景。
- 当前 domain 为 `unknown`：只保留低敏全局偏好。

### 第三阶段：结构升级

如果第一阶段效果稳定，再把 `ExtendedPreferences` 升级为分层结构，保留向后兼容：

```json
{
  "global_preferences": {
    "language_tone": "专业正式",
    "typical_page_count": 12,
    "content_tone": {"detail_level": 6}
  },
  "domain_profiles": {
    "business": {
      "templates": {"pitch-deck": 4},
      "themes": {"charcoal_light": 3},
      "layout_preferences": {"kpi_dashboard": 3}
    },
    "academic": {
      "templates": {"design-defense": 2},
      "themes": {"ocean_soft": 2}
    }
  }
}
```

## 风险与待确认问题

- 旧画像缺少每条偏好的来源 domain，只能用 `SuccessPatterns` 和 `DomainPreferences` 粗略判断，短期过滤可能会少用一些有效偏好。
- 当前有 web 层基础画像注入和 deep 层增强画像注入两条路径，需要确认任务创建时是否总会走 `ProcessUserIntent()`，否则统一注入点要兼容 fallback。
- `SpecialNotes` 可能混有用户手动编辑内容、LLM 总结和任务特定记忆，建议默认不跨领域注入。
- LLM Updater 分析信号时没有完整任务 domain/template/theme 上下文，后续需要在 `LearningSignal` 或 completion 回写中补齐。
- 用户如果明确说“沿用我以前的风格”，门控应允许显式 override。

## 结论

当前过拟合的主因是：历史偏好被当作全局强约束注入，而不是按当前任务相关性筛选后的弱参考。短期最有效的修复不是调模型参数，而是把“用户当前需求优先、历史偏好按字段敏感性和领域匹配降权”落到代码路径里。

已按 direct 小步事项实施第一阶段止血：统一偏好注入入口，增加 domain-aware 偏好筛选，调整 prompt 语气，并补充跨领域回归测试。后续如果需要继续提升个性化质量，再通过 OpenSpec 推进分层画像结构。

## 关联事项

- TODO: `docs/issues/todo.md#PPT-QUALITY-010`
- 现有设计参考：`ppt-agent/doc/learning-system-domain-match.md`
