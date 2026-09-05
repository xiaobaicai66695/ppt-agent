#!/usr/bin/env node
/**
 * Hydrate planned PPTSpec visuals with Unsplash images for an external Agent.
 *
 * Images are optional for a ppt. This CLI resolves planned backgrounds from
 * content_plan.visual_intent and foreground image components from
 * content_plan.components.
 */

import { mkdir, readFile, rename, stat, writeFile } from "node:fs/promises";
import { basename, extname, isAbsolute, join, relative, resolve } from "node:path";
import { fileURLToPath } from "node:url";

import { readAccessToken } from "./auth_store.ts";

const API_BASE_URL = "https://api.unsplash.com";
const MAX_IMAGE_BYTES = 20 * 1024 * 1024;
const DEFAULT_PER_PAGE = 3;

type JsonObject = Record<string, unknown>;

interface VisualIntent extends JsonObject {
  asset_purpose: string;
  asset_query: string;
  asset_subject: string;
  composition: string;
  orientation: "landscape" | "portrait" | "squarish";
  asset_id?: string;
  id?: string;
  local_path?: string;
  image_path?: string;
  asset_path?: string;
  path?: string;
  image_url?: string;
  preview_url?: string;
  source_url?: string;
  attribution?: string;
  provider?: string;
  search_status?: string;
  download_location?: string;
  photographer?: string;
  photographer_url?: string;
}

interface PlannedVisual {
  taskLabel: string;
  contentType: string;
  intent: VisualIntent;
}

type MaterializedAsset = Pick<VisualIntent, "asset_id" | "local_path" | "image_url" | "preview_url" | "source_url" | "attribution" | "provider" | "search_status">;

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

function cleanText(value: unknown): string {
  return typeof value === "string" ? value.trim() : "";
}

function firstPathValue(intent: VisualIntent): string {
  for (const key of ["local_path", "image_path", "asset_path", "path"] as const) {
    const value = cleanText(intent[key]);
    if (value) return value;
  }
  return "";
}

function firstQueryValue(intent: VisualIntent): string {
  return cleanText(intent.asset_query) || cleanText(intent.asset_subject);
}

function normalizePathAliases(intent: VisualIntent): void {
  const path = firstPathValue(intent);
  if (path && !cleanText(intent.local_path)) {
    intent.local_path = path;
  }
}

function normalizePlannedVisual(intent: VisualIntent, fieldPrefix: string, defaultPurpose: string): void {
  normalizePathAliases(intent);
  const hasPath = Boolean(firstPathValue(intent));
  const hasRemoteURL = Boolean(cleanText(intent.image_url));
  intent.asset_purpose = cleanText(intent.asset_purpose) || defaultPurpose;
  allowedAssetPurpose(intent.asset_purpose, `${fieldPrefix}.asset_purpose`);
  const query = firstQueryValue(intent);
  if (query) {
    intent.asset_query = query;
  } else if (!hasPath && !hasRemoteURL) {
    requiredText(intent.asset_query, `${fieldPrefix}.asset_query`);
  }
  intent.asset_subject = cleanText(intent.asset_subject) || cleanText(intent.asset_query);
  intent.composition = cleanText(intent.composition) || "wide landscape";
  const orientation = cleanText(intent.orientation) || "landscape";
  if (orientation !== "landscape" && orientation !== "portrait" && orientation !== "squarish") {
    fail(`${fieldPrefix}.orientation 必须是 landscape、portrait 或 squarish。`);
  }
  intent.orientation = orientation;
}

function visualIntentFor(task: JsonObject, taskLabel: string): VisualIntent | null {
  const contentPlan = asObject(task.content_plan, `${taskLabel}.content_plan`);
  if (contentPlan.visual_intent === undefined || contentPlan.visual_intent === null) return null;
  const intent = asObject(contentPlan.visual_intent, `${taskLabel}.content_plan.visual_intent`) as VisualIntent;
  normalizePlannedVisual(intent, `${taskLabel}.visual_intent`, "background");
  return intent;
}

function imageComponentVisualsFor(task: JsonObject, taskLabel: string): PlannedVisual[] {
  const contentPlan = asObject(task.content_plan, `${taskLabel}.content_plan`);
  if (contentPlan.components === undefined || contentPlan.components === null) return [];
  if (!Array.isArray(contentPlan.components)) fail(`${taskLabel}.content_plan.components 必须是数组。`);
  const visuals: PlannedVisual[] = [];
  for (let index = 0; index < contentPlan.components.length; index += 1) {
    const component = asObject(contentPlan.components[index], `${taskLabel}.content_plan.components[${index}]`) as VisualIntent;
    if (cleanText(component.type) !== "image") continue;
    const componentLabel = `${taskLabel}.content_plan.components[${index}]`;
    normalizePlannedVisual(component, componentLabel, "scene");
    visuals.push({ taskLabel: componentLabel, contentType: contentTypeKey(task), intent: component });
  }
  return visuals;
}

function allowedAssetPurpose(value: string, field: string): void {
  if (!new Set(["background", "scene", "evidence", "decorative"]).has(value)) {
    fail(`${field} 必须是 background、scene、evidence 或 decorative。`);
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
      "User-Agent": "ppt-planner/unsplash",
    },
  });
  if (!response.ok) {
    if (response.status === 401) fail("Unsplash Access Key 无效或未获授权。请在 Unsplash Developers 检查应用权限后重新运行 unsplash auth。");
    fail(`${operation} 失败：Unsplash 返回 HTTP ${response.status}。`);
  }
  return await response.json() as T;
}

async function searchPhoto(intent: VisualIntent, key: string, perPage: number): Promise<UnsplashPhoto> {
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

async function downloadPhoto(photo: UnsplashPhoto, workDir: string, key: string): Promise<Pick<VisualIntent, "asset_id" | "local_path" | "image_url" | "preview_url" | "source_url" | "attribution">> {
  const imageUrl = photo.urls.regular ?? photo.urls.full ?? photo.urls.small;
  if (!imageUrl || !isUnsplashUrl(imageUrl, false)) fail("Unsplash 返回了不可信的图片地址。");
  if (photo.links.download_location) {
    if (!isUnsplashUrl(photo.links.download_location, true)) fail("Unsplash 返回了不可信的下载追踪地址。");
    await requestJson<JsonObject>(photo.links.download_location, key, "图片下载追踪");
  }
  const imageResponse = await fetch(imageUrl, { headers: { "User-Agent": "ppt-planner/unsplash" } });
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
    asset_id: photo.id,
    local_path: relative(workDir, output).replaceAll("\\", "/"),
    image_url: imageUrl,
    preview_url: photo.urls.small ?? imageUrl,
    source_url: photo.links.html ?? "",
    attribution: photographer ? `Photo by ${photographer} on Unsplash` : "Photo on Unsplash",
  };
}

async function hasResolvedImage(intent: VisualIntent, workDir: string): Promise<boolean> {
  normalizePathAliases(intent);
  if (intent.provider !== "unsplash" || intent.search_status !== "resolved" || !intent.local_path || !intent.source_url || !intent.attribution) return false;
  const path = safeWorkPath(workDir, intent.local_path);
  try {
    return (await stat(path)).isFile();
  } catch {
    return false;
  }
}

async function hasLocalImage(intent: VisualIntent, workDir: string): Promise<boolean> {
  normalizePathAliases(intent);
  if (!intent.local_path) return false;
  const path = safeWorkPath(workDir, intent.local_path);
  try {
    return (await stat(path)).isFile();
  } catch {
    return false;
  }
}

function remotePhotoFromIntent(intent: VisualIntent): UnsplashPhoto | null {
  const imageUrl = cleanText(intent.image_url);
  if (!imageUrl) return null;
  if (!isUnsplashUrl(imageUrl, false)) fail("image_url 必须是可信的 Unsplash 图片地址。");
  const previewUrl = cleanText(intent.preview_url);
  if (previewUrl && !isUnsplashUrl(previewUrl, false)) fail("preview_url 必须是可信的 Unsplash 图片地址。");
  const downloadLocation = cleanText(intent.download_location);
  if (downloadLocation && !isUnsplashUrl(downloadLocation, true)) fail("download_location 必须是可信的 Unsplash API 地址。");
  return {
    id: cleanText(intent.asset_id) || cleanText(intent.id) || "selected",
    width: 1,
    height: 1,
    urls: {
      regular: imageUrl,
      small: previewUrl || imageUrl,
    },
    links: {
      html: cleanText(intent.source_url),
      download_location: downloadLocation,
    },
    user: {
      name: cleanText(intent.photographer),
      links: { html: cleanText(intent.photographer_url) },
    },
  };
}

function resolvedMetadata(intent: VisualIntent, photo: UnsplashPhoto, downloaded: Awaited<ReturnType<typeof downloadPhoto>>): MaterializedAsset {
  return {
    ...downloaded,
    source_url: downloaded.source_url || cleanText(intent.source_url),
    attribution: downloaded.attribution === "Photo on Unsplash" && cleanText(intent.attribution) ? cleanText(intent.attribution) : downloaded.attribution,
    asset_id: downloaded.asset_id || cleanText(intent.asset_id) || cleanText(intent.id) || photo.id,
    provider: "unsplash",
    search_status: "resolved",
  };
}

async function hydrateStandaloneVisual(taskLabel: string, intent: VisualIntent, workDir: string, perPage: number, completed: string[]): Promise<void> {
  if (await hasLocalImage(intent, workDir)) {
    completed.push(`${taskLabel}: 已存在`);
    return;
  }
  const accessToken = await readAccessToken();
  const remotePhoto = remotePhotoFromIntent(intent);
  if (remotePhoto) {
    const downloaded = await downloadPhoto(remotePhoto, workDir, accessToken);
    Object.assign(intent, resolvedMetadata(intent, remotePhoto, downloaded));
    completed.push(`${taskLabel}: 已下载已选 Unsplash 图片 ${intent.asset_id}`);
    return;
  }
  const query = firstQueryValue(intent);
  if (!query) {
    fail(`${taskLabel}.asset_query 是必填项。请先提供 asset_query，或提供可下载的 image_url。`);
  }
  intent.asset_query = query;
  const photo = await searchPhoto(intent, accessToken, perPage);
  const downloaded = await downloadPhoto(photo, workDir, accessToken);
  Object.assign(intent, resolvedMetadata(intent, photo, downloaded));
  completed.push(`${taskLabel}: 已下载 unsplash_${photo.id}`);
}

function isBackgroundIntent(intent: VisualIntent): boolean {
  const purpose = String(intent.asset_purpose || "").trim().toLowerCase();
  const role = String(intent.role || "").trim().toLowerCase();
  const position = String(intent.image_position || "").trim().toLowerCase();
  return purpose === "background" || role === "background" || role === "hero_photo" || position === "background";
}

function contentTypeKey(task: JsonObject): string {
  const contentType = String(task.content_type || "").trim().toLowerCase();
  return contentType || "content_slide";
}

function materializedAsset(intent: VisualIntent): MaterializedAsset {
  return {
    asset_id: intent.asset_id,
    local_path: intent.local_path,
    image_url: intent.image_url,
    preview_url: intent.preview_url,
    source_url: intent.source_url,
    attribution: intent.attribution,
    provider: intent.provider,
    search_status: intent.search_status,
  };
}

function applySharedBackground(intent: VisualIntent, source: VisualIntent): void {
  Object.assign(intent, materializedAsset(source), {
    asset_query: source.asset_query,
    orientation: source.orientation,
  });
}

async function hydrateSharedBackgroundGroup(group: PlannedVisual[], workDir: string, key: string, perPage: number, completed: string[]): Promise<void> {
  const reusable = await findResolvedVisual(group, workDir);
  const source = reusable ?? group[0];
  if (!reusable) {
    const accessToken = await readAccessToken();
    const remotePhoto = remotePhotoFromIntent(source.intent);
    const photo = remotePhoto ?? await searchPhoto(source.intent, accessToken, perPage);
    const downloaded = await downloadPhoto(photo, workDir, accessToken);
    Object.assign(source.intent, resolvedMetadata(source.intent, photo, downloaded));
    const sourceLabel = remotePhoto ? `已下载已选 Unsplash 图片 ${source.intent.asset_id}` : `已下载 unsplash_${photo.id}`;
    completed.push(`${source.taskLabel}: ${sourceLabel}（${key} 共享背景）`);
  } else {
    completed.push(`${source.taskLabel}: 已存在（${key} 共享背景）`);
  }
  for (const target of group) {
    applySharedBackground(target.intent, source.intent);
    if (target !== source) {
      completed.push(`${target.taskLabel}: 复用 ${source.taskLabel} 的 ${key} 背景`);
    }
  }
}

async function findResolvedVisual(group: PlannedVisual[], workDir: string): Promise<PlannedVisual | null> {
  for (const planned of group) {
    if (await hasResolvedImage(planned.intent, workDir)) {
      return planned;
    }
  }
  return null;
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
  if (!Array.isArray(manifest.tasks) || !manifest.tasks.length) fail("tasks.json 必须包含非空 tasks 数组。");

  const plannedVisuals: PlannedVisual[] = [];
  const skipped: string[] = [];
  for (let index = 0; index < manifest.tasks.length; index += 1) {
    const task = asObject(manifest.tasks[index], `tasks[${index}]`);
    const taskLabel = String(task.task_id || task.page_index || index + 1);
    const intent = visualIntentFor(task, `tasks[${index}](${taskLabel})`);
    if (intent) {
      allowedAssetPurpose(intent.asset_purpose, `tasks[${index}](${taskLabel}).visual_intent.asset_purpose`);
      plannedVisuals.push({ taskLabel, contentType: contentTypeKey(task), intent });
    }
    const imageComponents = imageComponentVisualsFor(task, `tasks[${index}](${taskLabel})`);
    plannedVisuals.push(...imageComponents);
    if (!intent && !imageComponents.length) skipped.push(taskLabel);
  }
  if (!plannedVisuals.length) {
    console.log(JSON.stringify({ ok: true, work_dir: options.workDir, assets: [], skipped_no_visual_intent: skipped }, null, 2));
    return;
  }

  const completed: string[] = [];
  const backgroundGroups = new Map<string, PlannedVisual[]>();
  for (const planned of plannedVisuals) {
    if (!isBackgroundIntent(planned.intent)) continue;
    const group = backgroundGroups.get(planned.contentType) ?? [];
    group.push(planned);
    backgroundGroups.set(planned.contentType, group);
  }
  for (const [contentType, group] of backgroundGroups) {
    await hydrateSharedBackgroundGroup(group, options.workDir, contentType, options.perPage, completed);
  }

  for (const { taskLabel, intent } of plannedVisuals) {
    if (isBackgroundIntent(intent)) continue;
    await hydrateStandaloneVisual(taskLabel, intent, options.workDir, options.perPage, completed);
  }
  await writeFile(manifestPath, `${JSON.stringify(manifest, null, 2)}\n`, "utf8");
  console.log(JSON.stringify({ ok: true, work_dir: options.workDir, assets: completed }, null, 2));
}

const invokedDirectly = process.argv[1] && resolve(process.argv[1]) === fileURLToPath(import.meta.url);
if (invokedDirectly) {
  hydrateUnsplashAssets(process.argv.slice(2)).catch((error) => {
    console.error(`图片资产处理失败: ${error instanceof Error ? error.message : String(error)}`);
    process.exit(1);
  });
}
