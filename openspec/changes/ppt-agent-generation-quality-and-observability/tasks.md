## 1. 推荐与动态规划

- [x] 1.1 让推荐模式复用一次结构化 LLM 意图结果，校验建议模板和建议页数并提供通用降级值
- [x] 1.2 新增 recommended-style outline 语义，只传视觉风格与页数提示，不复制 preset 的 DefaultSlides
- [x] 1.3 保持显式 preset 的 template-scaffold 行为并补充推荐/显式模式回归测试

## 2. 背景与内容容量

- [x] 2.1 为主 Agent 提供真实背景候选、页面语义优先级和整套 45%-65% 覆盖目标
- [x] 2.2 在 manifest 更新链路校验背景引用并为推荐模式提供低覆盖率确定性补齐与同主题图片轮换
- [x] 2.3 放宽信息页模板 contract，将目标密度与最大渲染容量分离并同步 generator 参考文档
- [x] 2.4 重写相关 prompt 为正向目标、决策顺序和结构化容量契约，补充 prompt/模板校验测试

## 3. Runtime Metadata 与工具轨迹

- [x] 3.1 将模型状态注入和 manifest_validated 事件 metadata 收敛为 done/total
- [x] 3.2 确认所有 callback ToolCall 类型持久化，保持完整历史独立于 RecentEvents 热窗口
- [x] 3.3 在前端按稳定事件身份合并会话历史与 SSE recent_events，补充任务切换和重复事件测试

## 4. 上下文压缩可观测性

- [x] 4.1 提取首个用户请求与后续明确约束，生成结构化压缩 handoff 和确定性失败兜底
- [x] 4.2 扩展压缩事件 metadata，记录消息数、token 数、节省量/比例及有界保留要求
- [x] 4.3 新增前端 compression 分类和专用前后对比详情，并兼容旧事件字段

## 5. 验证与交付

- [x] 5.1 运行后端聚焦测试、go test ./... 和 go build ./...
- [x] 5.2 运行前端单元测试与 npm run build，并执行单页模板 JSON/容量校验
- [x] 5.3 执行 openspec validate --strict，更新迭代记录与 done.md 归档
- [x] 5.4 部署到 Linux 服务器，重启服务并验证健康接口、模板接口和关键生成链路
