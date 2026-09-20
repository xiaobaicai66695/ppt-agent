# 2026-09-20 newppt.xyz Nginx 反向代理与 TLS 配置记录

## 已完成

- 部署目标：`remote-dev:/ppt/ppt-agent`，应用服务 `ppt-agent.service` 保持监听 `:8080`，未中断现有直接访问方式。
- 确认 `newppt.xyz` 的 A 记录解析至 `124.220.22.162`，公网 HTTP 访问 `http://newppt.xyz/api/health` 返回 301，目标为 HTTPS 地址。
- 使用现有 `nginx:alpine` Docker 镜像启动 `ppt-agent-nginx` 容器，映射宿主机 `80:80`、`443:443`，反向代理到 `host.docker.internal:8080`；SSE 已禁用代理缓冲并设置长读写超时。
- Let’s Encrypt 已为 `newppt.xyz` 签发证书，证书主体为 `CN = newppt.xyz`，有效期至 2026-12-19。
- HTTPS 配置启用了 TLS 1.2/1.3、HTTP 到 HTTPS 重定向和 HSTS。`ppt-agent-certbot-renew.timer` 已启用，按日执行续期检查并在执行后 reload Nginx。

## 验收与遗留项

- 服务端本机 HTTPS：`curl --resolve newppt.xyz:443:127.0.0.1 https://newppt.xyz/api/health` 返回 `{"status":"ok"}`；Nginx 配置检查通过；`http://127.0.0.1:8080/api/health` 仍返回 200。
- 当前云安全组尚未开放 TCP 443：公网直连 `https://newppt.xyz/api/health` 超时，不能视为 HTTPS 已完成外网闭环。需在云控制台新增入站 `TCP 443`（来源按业务策略，公开网站通常为 `0.0.0.0/0`），随后复测域名 HTTPS。
- TCP 80 必须保留开放，供 Let’s Encrypt 的 HTTP-01 续期校验使用。TCP 8080 目前按用户需求继续暴露；若不再需要直连，后续应关闭安全组 8080 并将服务绑定为 `127.0.0.1:8080`。
