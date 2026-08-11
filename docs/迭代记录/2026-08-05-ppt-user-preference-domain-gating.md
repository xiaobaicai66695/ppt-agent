# PPT 用户偏好领域门控迭代记录

## 背景

用户反馈：PPT Agent 的用户偏好学习存在跨场景过拟合，用户输入与历史常用场景差异较大时，系统仍会按旧场景偏好填充 `tasks.json`，尤其是模板、主题、配色、布局和内容叙事。

## 本轮目标

- 补充成熟 PPT 生成产品的偏好/品牌约束做法，形成预研结论。
- 将历史偏好从“全局强约束”降级为“按领域命中的弱参考”。
- 保留同领域个性化收益，同时阻断跨领域模板、主题、配色、布局和备注迁移。
- 补充回归测试，覆盖跨领域不注入和同领域可复用两类关键路径。

## 实现摘要

- `ppt-agent/backend/pkg/style/profile.go`
  - 新增 `GetPreferredTemplatesForDomain()`、`GetPreferredThemeForDomain()`、`HasExactDomainHistory()`。
  - 只有当前领域存在精确历史证据时，才返回模板和主题这类场景敏感偏好。
  - 跳过空领域和 `unknown` 领域的成功模式学习，减少无效画像强化。

- `ppt-agent/backend/pkg/agent/router/engine.go`
  - `ProfileMatcher.EnhanceWithProfile()` 不再无条件前置历史模板。
  - 同领域历史才参与模板排序和主题补位。
  - 典型页数只在当前分类未给出页数估计时补位，不再和当前估计取平均。

- `ppt-agent/backend/pkg/agent/deck/agent.go`
  - `ProcessUserIntent()` 将意图识别出的 domain 传入偏好注入逻辑。
  - `enhanceStyleContextWithProfile()` 改为“用户偏好参考”文案，明确当前主题、用户显式大纲、显式模板/配色优先。
  - 语言风格、常用页数、动画、图表等低敏字段可跨领域弱参考；模板、主题、配色、布局、备注、成功经验等高敏字段必须同领域命中。

- `ppt-agent/backend/pkg/web/handler.go`
  - 移除任务创建早期的 `BuildStyleContext()` 注入，避免还没完成 intent/domain 判断时就把全局历史画像塞进主 prompt。

- `ppt-agent/backend/pkg/task/manager.go`
  - 完成任务学习时补齐 domain、template、theme、page_count，使后续成功模式具备领域证据。

- `ppt-agent/backend/pkg/agent/learning/engine.go`
  - 忽略空领域和 `unknown` 领域的领域偏好、成功模式记录。

## 调研补充

预研文档已补充市面成熟产品做法，包括 Microsoft Copilot for PowerPoint、Canva、Gamma、Beautiful.ai、Pitch、Presentations.AI。共性结论是：成熟产品通常把品牌/模板作为显式选择或结构化资产，当前 prompt、文件、URL、模板优先，历史学习结果只作为辅助信号。

文档：`docs/research/2026-08-05-ppt-user-preference-overfit.md`

## 验证

- `go test ./pkg/agent/router ./pkg/agent/deck ./pkg/agent/learning ./pkg/style ./pkg/task ./pkg/web`
- `go build ./...`
- `go test ./...`

新增测试：

- `ppt-agent/backend/pkg/agent/router/profile_matcher_test.go`
  - 跨领域历史不前置模板、不覆盖主题、不影响已识别页数。
  - 同领域历史可补充模板和主题。
- `ppt-agent/backend/pkg/agent/deck/preference_context_test.go`
  - 跨领域不注入模板、主题、配色、布局、备注等高敏感偏好。
  - 同领域可注入场景敏感偏好。

## 上线

- 本地构建 Linux 二进制：`ppt-agent/backend/ppt-agent-linux.new`
- 已复制到远端并替换：`/ppt/ppt-agent/backend/ppt-agent-linux`
- 旧二进制备份：`/ppt/ppt-agent/backend/ppt-agent-linux.bak.20260805135822`
- 已重启远端服务：
  - 进程 PID：`427674`
  - 健康检查：`/api/health` 返回 `{"status":"ok"}`
  - 模板接口：`/api/templates` 正常返回 preset 列表

## 后续建议

- 中期把画像升级为 `global_preferences` + `domain_profiles`，将模板、主题、配色、布局、品牌资产从自然语言 prompt 迁移到结构化 profile bucket。
- 前端可增加“沿用历史风格/当前品牌包”的显式选择，让用户能主动打开跨领域复用。
- 增加线上可观测字段，记录每次生成使用了哪些偏好来源：当前显式选择、同领域历史、全局弱偏好或未使用。

## 补充：确定性用户资料直注

根据后续反馈，用户画像中的姓名、工作单位、部门、职位、行业、地区等确定性资料不属于历史风格偏好，不需要跟随 domain-aware 门控。已新增 `user_facts` 扩展字段，存储在 `extended_preferences` 中，并在主 Agent `StyleContext` 中以“用户确定性资料”独立注入。

规则：

- `user_facts` 可直接作为称谓、署名、组织背景和工作场景上下文。
- 如果当前任务显式给出了不同单位、身份或署名，以当前任务为准。
- 模板、主题、配色、布局、历史备注等风格偏好仍然保持同领域门控。
