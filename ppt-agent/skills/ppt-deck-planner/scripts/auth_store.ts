import { chmod, readFile, writeFile } from "node:fs/promises";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";

export const AUTH_FILE = resolve(dirname(fileURLToPath(import.meta.url)), "..", "auth.txt");
const ACCESS_KEY_ENV_NAMES = ["UNSPLASH_ACCESS_KEY", "UNSPLASH_ACCESS_TOKEN"] as const;

export async function saveAccessToken(accessToken: string): Promise<void> {
  const value = accessToken.trim();
  if (!value || /\s/.test(value)) {
    throw new Error("Access Key 不能为空且不能包含空白字符。");
  }
  await writeFile(AUTH_FILE, `${value}\n`, { encoding: "utf8", mode: 0o600 });
  try {
    await chmod(AUTH_FILE, 0o600);
  } catch {
    // Windows does not implement POSIX permissions; the file remains ignored by Git.
  }
}

export async function readAccessToken(): Promise<string> {
	try {
		const value = (await readFile(AUTH_FILE, "utf8")).trim();
		if (!value || /\s/.test(value)) {
			throw new Error("auth.txt 中的 Access Key 无效，请重新执行 unsplash auth。");
		}
		return value;
	} catch (error) {
		if (error instanceof Error && error.message.includes("auth.txt")) {
			throw error;
		}
		return readAccessTokenFromEnv();
	}
}

export function readAccessTokenFromEnv(env: NodeJS.ProcessEnv = process.env): string {
	for (const name of ACCESS_KEY_ENV_NAMES) {
		const value = env[name]?.trim() ?? "";
		if (!value) continue;
		if (/\s/.test(value)) {
			throw new Error(`${name} 不能为空且不能包含空白字符。`);
		}
		return value;
	}
	throw new Error(`未找到 auth.txt，也未配置 ${ACCESS_KEY_ENV_NAMES.join(" 或 ")}。请执行 unsplash auth 或 unsplash auth --from-env。`);
}
