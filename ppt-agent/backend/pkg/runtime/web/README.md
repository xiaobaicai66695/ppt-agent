# Web handlers

Web 代码按业务职责导航。所有实现文件都在当前目录并属于同一 Go package，以保留 `Server` 未导出依赖；不要创建只有 README 的子目录。

- `auth/`：登录、注册、会话鉴权
- `task/`：任务创建、查询、取消、反馈
- `conversation/`：对话与继续任务
- `delivery/`：SSE、下载、缩略图
- `admin/`：凭据、日志分析、管理接口
