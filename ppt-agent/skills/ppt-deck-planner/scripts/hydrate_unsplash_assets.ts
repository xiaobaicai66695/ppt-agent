#!/usr/bin/env node
/**
 * Hydrate planned DeckSpec visuals with Unsplash images for an external Agent.
 *
 * The CLI resolves every planned page visual_intent and foreground image
 * component. It never writes a partial manifest: all metadata is persisted
 * only after the planned asset set has been resolved.
 */

import { mkdir, readFile, rename, stat, writeFile } from "node:fs/promises";
import { basename, extname, isAbsolute, join, relative, resolve } from "node:path";
import { fileURLToPath } from "node:url";

import { readAccessToken } from "./auth_store.ts";

const API_BASE_URL = "https://api.unsplash.com";
const MAX_IMAGE_BYTES = 20 * 1024 * 1024;
const DEFAULT_PER_PAGE = 3;

type JsonObject = Record<string, unknown>;

interface VisualAsset extends JsonObject {
  asset_purpose: string;
  asset_query: string;
  asset_subject: string;
  composition: string;
  orientation: "landscape" | "portrait" | "squarish";
  local_path?: string;
  image_url?: string;
  preview_url?: string;
  source_url?: string;
  attribution?: string;
  provider?: string;
  search_status?: string;
}

interface UnsplashPhoto {
  id: string;
  width: number;
  height: number;
  urls: { regular?: string; small?: string; full?: string };
  links: { html?: string; download_location?: string };
  user: { name?: string; username?: string; links?: { html?: string } };
}

interface CliOptions {
  workDir: string;
  perPage: number;
  externalAgent: boolean;
}

function fail(message: string): never {
  console.error(message);
  process.exit(1);
}

function parseArgs(argv: string[]): CliOptions {
  let workDir = "";
  let perPage = DEFAULT_PER_PAGE;
  let externalAgent = false;
  for (let index = 0; index < argv.length; index += 1) {
    const value = argv[index];
    if (value === "--work-dir") {
      workDir = argv[index + 1] ?? "";
      index += 1;
    } else if (value === "--per-page") {
      perPage = Number.parseInt(argv[index + 1] ?? "", 10);
      index += 1;
    } else if (value === "--external-agent") {
      externalAgent = true;
    } else if (value === "--help" || value === "-h") {
      console.log("Usage: unsplash fetch --work-dir <directory> [--per-page 1-10]");
      process.exit(0);
    } else {
      fail(`未知参数: ${value}`);
    }
  }
  if (!workDir) {
    fail("缺少 --work-dir。用法：unsplash fetch --work-dir <包含 tasks.json 的目录>");
  }
  if (!Number.isInteger(perPage) || perPage < 1 || perPage > 10) {
    fail("--per-page 必须是 1 到 10 的整数。");
  }
  if (!externalAgent) {
    fail("该图片脚本只供 ppt-agent 项目外的 Agent 使用。请使用 --external-agent 明确确认，项目内请使用后端既有素材链路。");
  }
  return { workDir: resolve(workDir), perPage, externalAgent };
}

function asObject(value: unknown, field: string): JsonObject {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    fail(`${field} 必须是对象。`);
  }
  return value as JsonObject;
}

function requiredText(value: unknown, field: string): string {
  if (typeof value !== "string" || !value.trim()) {
    fail(`${field} 是必填项。请先完成页面视觉规划，再运行图片下载脚本。`);
  }
  return value.trim();
}

function visualAssetFor(value: unknown, label: string, defaultPurpose: string): VisualAsset {
  const asset = asObject(value, label) as VisualAsset;
  asset.asset_purpose = typeof asset.asset_purpose === "string" && asset.asset_purpose.trim()
    ? asset.asset_purpose.trim()
    : defaultPurpose;
  asset.asset_query = requiredText(asset.asset_query, `${label}.asset_query`);
  asset.asset_subject = requiredText(asset.asset_subject, `${label}.asset_subject`);
  asset.composition = requiredText(asset.composition, `${label}.composition`);
  const orientation = typeof asset.orientation === "string" && asset.orientation.trim()
    ? asset.orientation.trim()
    : "landscape";
  if (orientation !== "landscape" && orientation !== "portrait" && orientation !== "squarish") {
    fail(`${label}.orientation 必须是 landscape、portrait 或 squarish。`);
  }
  asset.orientation = orientation as VisualAsset["orientation"];
  return asset;
}

function allowedAssetPurpose(value: string, taskLabel: string): void {
  if (!new Set(["background", "scene", "evidence", "decorative"]).has(value)) {
    fail(`${taskLabel}.visual_intent.asset_purpose 必须是 background、scene、evidence 或 decorative。`);
  }
}

function safeWorkPath(workDir: string, candidate: string): string {
  const target = resolve(workDir, candidate);
  const rel = relative(workDir, target);
  if (rel === "" || rel === ".." || rel.startsWith("..\\") || rel.startsWith("../") || isAbsolute(rel)) {
    fail(`图片路径必须位于工作目录内: ${candidate}`);
  }
  return target;
}

function isUnsplashUrl(value: string, api: boolean): boolean {
  try {
    const url = new URL(value);
    if (url.protocol !== "https:") return false;
    const host = url.hostname.toLowerCase();
    return api ? host === "api.unsplash.com" : host === "unsplash.com" || host.endsWith(".unsplash.com");
  } catch {
    return false;
  }
}

async function requestJson<T>(url: string, key: string, operation: string): Promise<T> {
  const response = await fetch(url, {
    headers: {
      Authorization: `Client-ID ${key}`,
      "Accept-Version": "v1",
      Accept: "application/json",
      "User-Agent": "ppt-deck-planner/unsplash",
    },
  });
  if (!response.ok) {
    if (response.status === 401) fail("Unsplash Access Key 无效或未获授权。请在 Unsplash Developers 检查应用权限后重新运行 unsplash auth。");
    fail(`${operation} 失败：Unsplash 返回 HTTP ${response.status}。`);
  }
  return await response.json() as T;
}

async function searchPhoto(intent: VisualAsset, key: string, perPage: number): Promise<UnsplashPhoto> {
  const url = new URL("/search/photos", API_BASE_URL);
  url.searchParams.set("query", intent.asset_query);
  url.searchParams.set("orientation", intent.orientation);
  url.searchParams.set("content_filter", "high");
  url.searchParams.set("order_by", "relevant");
  url.searchParams.set("page", "1");
  url.searchParams.set("per_page", String(perPage));
  const payload = await requestJson<{ results?: UnsplashPhoto[] }>(url.toString(), key, "图片搜索");
  const candidates = (payload.results ?? []).filter((photo) => photo.id && photo.width > 0 && photo.height > 0 && (photo.urls.regular || photo.urls.full || photo.urls.small));
  if (!candidates.length) fail(`图片搜索未返回可下载结果：${intent.asset_query}`);
  const targetRatio = intent.orientation === "landscape" ? 16 / 9 : intent.orientation === "portrait" ? 3 / 4 : 1;
  return [...candidates].sort((left, right) => scorePhoto(right, targetRatio) - scorePhoto(left, targetRatio))[0];
}

function scorePhoto(photo: UnsplashPhoto, targetRatio: number): number {
  const ratio = photo.width / photo.height;
  const ratioScore = 1 / (1 + Math.abs(ratio - targetRatio));
  return ratioScore * 10 + Math.log10(photo.width * photo.height);
}

function extensionFor(contentType: string, sourceUrl: string): string {
  const normalized = contentType.toLowerCase().split(";", 1)[0].trim();
  if (normalized === "image/jpeg") return ".jpg";
  if (normalized === "image/png") return ".png";
  if (normalized === "image/webp") return ".webp";
  const suffix = extname(new URL(sourceUrl).pathname).toLowerCase();
  if ([".jpg", ".jpeg", ".png", ".webp"].includes(suffix)) return suffix;
  fail(`Unsplash 返回了不受支持的图片类型: ${contentType || "未知"}`);
}

async function downloadPhoto(photo: UnsplashPhoto, workDir: string, key: string): Promise<Pick<VisualAsset, "local_path" | "image_url" | "preview_url" | "source_url" | "attribution">> {
  const imageUrl = photo.urls.regular ?? photo.urls.full ?? photo.urls.small;
  if (!imageUrl || !isUnsplashUrl(imageUrl, false)) fail("Unsplash 返回了不可信的图片地址。");
  if (photo.links.download_location) {
    if (!isUnsplashUrl(photo.links.download_location, true)) fail("Unsplash 返回了不可信的下载追踪地址。");
    await requestJson<JsonObject>(photo.links.download_location, key, "图片下载追踪");
  }
  const imageResponse = await fetch(imageUrl, { headers: { "User-Agent": "ppt-deck-planner/unsplash" } });
  if (!imageResponse.ok) fail(`图片下载失败：Unsplash 返回 HTTP ${imageResponse.status}。`);
  if (!isUnsplashUrl(imageResponse.url, false)) fail("图片下载被重定向到不可信地址。");
  const bytes = new Uint8Array(await imageResponse.arrayBuffer());
  if (!bytes.byteLength || bytes.byteLength > MAX_IMAGE_BYTES) fail("图片下载为空或超过 20MB 限制。");
  const extension = extensionFor(imageResponse.headers.get("content-type") ?? "", imageResponse.url);
  const assetDir = join(workDir, "assets", "images");
  await mkdir(assetDir, { recursive: true });
  const output = join(assetDir, `unsplash_${photo.id.replace(/[^a-zA-Z0-9_-]/g, "_")}${extension}`);
  const temp = join(assetDir, `.${basename(output)}.tmp`);
  await writeFile(temp, bytes);
  await rename(temp, output);
  const photographer = photo.user.name?.trim() || photo.user.username?.trim();
  return {
    local_path: relative(workDir, output).replaceAll("\\", "/"),
    image_url: imageUrl,
    preview_url: photo.urls.small ?? imageUrl,
    source_url: photo.links.html ?? "",
    attribution: photographer ? `Photo by ${photographer} on Unsplash` : "Photo on Unsplash",
  };
}

async function hasReadableLocalPath(asset: VisualAsset, workDir: string): Promise<boolean> {
  if (!asset.local_path) return false;
  const path = safeWorkPath(workDir, asset.local_path);
  try {
    return (await stat(path)).isFile();
  } catch {
    return false;
  }
}

async function hasResolvedImage(asset: VisualAsset, workDir: string): Promise<boolean> {
  if (!await hasReadableLocalPath(asset, workDir)) return false;
  return asset.provider === "unsplash" ? asset.search_status === "resolved" && !!asset.source_url && !!asset.attribution : true;
}

interface PlannedAsset {
  label: string;
  asset: VisualAsset;
}

async function collectPlannedAssets(manifest: JsonObject, workDir: string): Promise<{ planned: PlannedAsset[]; skipped: string[]; completed: string[] }> {
  if (!Array.isArray(manifest.tasks) || !manifest.tasks.length) fail("tasks.json 必须包含非空 tasks 数组。");
  const planned: PlannedAsset[] = [];
  const skipped: string[] = [];
  const completed: string[] = [];
  for (let index = 0; index < manifest.tasks.length; index += 1) {
    const task = asObject(manifest.tasks[index], `tasks[${index}]`);
    const taskLabel = String(task.task_id || task.page_index || index + 1);
    const plan = asObject(task.content_plan, `tasks[${index}](${taskLabel}).content_plan`);
    if (plan.visual_intent && asObject(plan.visual_intent, `${taskLabel}.visual_intent`).role !== "clean_text_only") {
      const existing = asObject(plan.visual_intent, `${taskLabel}.visual_intent`) as VisualAsset;
      if (await hasReadableLocalPath(existing, workDir)) completed.push(`${taskLabel}: 背景已存在`);
      else {
        const asset = visualAssetFor(plan.visual_intent, `tasks[${index}](${taskLabel}).content_plan.visual_intent`, "background");
        planned.push({ label: `${taskLabel}: 背景`, asset });
      }
    } else {
      skipped.push(taskLabel);
    }
    const components = Array.isArray(plan.components) ? plan.components : [];
    for (let componentIndex = 0; componentIndex < components.length; componentIndex += 1) {
      const component = asObject(components[componentIndex], `tasks[${index}].content_plan.components[${componentIndex}]`);
      if (component.type !== "image") continue;
      const existing = component as VisualAsset;
      if (await hasReadableLocalPath(existing, workDir)) completed.push(`${taskLabel}: 前景图已存在`);
      else {
        const asset = visualAssetFor(component, `tasks[${index}](${taskLabel}).content_plan.components[${componentIndex}]`, "scene");
        planned.push({ label: `${taskLabel}: 前景图`, asset });
      }
    }
  }
  return { planned, skipped, completed };
}

export async function hydrateUnsplashAssets(argv: string[]): Promise<void> {
  const options = parseArgs(argv);
  const manifestPath = join(options.workDir, "tasks.json");
  let manifest: JsonObject;
  try {
    manifest = JSON.parse(await readFile(manifestPath, "utf8")) as JsonObject;
  } catch (error) {
    fail(`无法读取 ${manifestPath}: ${error instanceof Error ? error.message : String(error)}`);
  }
  const targets = await collectPlannedAssets(manifest, options.workDir);
  if (!targets.planned.length) {
    console.log(JSON.stringify({ ok: true, work_dir: options.workDir, assets: targets.completed, skipped_no_visual_intent: targets.skipped }, null, 2));
    return;
  }

  const key = await readAccessToken();
  for (const { label, asset } of targets.planned) {
    allowedAssetPurpose(asset.asset_purpose, label);
    const photo = await searchPhoto(asset, key, options.perPage);
    Object.assign(asset, await downloadPhoto(photo, options.workDir, key), {
      provider: "unsplash",
      search_status: "resolved",
    });
    targets.completed.push(`${label}: 已下载 unsplash_${photo.id}`);
  }
  await writeFile(manifestPath, `${JSON.stringify(manifest, null, 2)}\n`, "utf8");
  console.log(JSON.stringify({ ok: true, work_dir: options.workDir, assets: targets.completed }, null, 2));
}

const invokedDirectly = process.argv[1] && resolve(process.argv[1]) === fileURLToPath(import.meta.url);
if (invokedDirectly) {
  hydrateUnsplashAssets(process.argv.slice(2)).catch((error) => {
    console.error(`图片资产处理失败: ${error instanceof Error ? error.message : String(error)}`);
    process.exit(1);
  });
}
