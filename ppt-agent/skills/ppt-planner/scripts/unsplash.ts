#!/usr/bin/env node
import { createInterface } from "node:readline";

import { AUTH_FILE, readAccessTokenFromEnv, saveAccessToken } from "./auth_store.ts";
import { hydrateUnsplashAssets } from "./hydrate_unsplash_assets.ts";

function usage(): void {
  console.log("Usage: unsplash <auth|fetch> [options]");
	console.log("  unsplash auth [--from-env]");
  console.log("  unsplash fetch --work-dir <directory> [--per-page 1-10]");
}

async function readSecret(prompt: string): Promise<string> {
  if (!process.stdin.isTTY || !process.stdout.isTTY) {
    throw new Error("unsplash auth 必须在交互式终端中运行。");
  }
  const readline = createInterface({ input: process.stdin, output: process.stdout, terminal: true });
  const originalWrite = readline._writeToOutput.bind(readline);
  let writingPrompt = true;
  readline._writeToOutput = (value: string) => {
    if (writingPrompt) {
      writingPrompt = false;
      originalWrite(value);
      return;
    }
    if (value.includes("\n") || value.includes("\r")) {
      originalWrite(value);
    }
  };
  return await new Promise<string>((resolve, reject) => {
    readline.question(prompt, (answer) => {
      readline.close();
      process.stdout.write("\n");
      resolve(answer);
    });
    readline.once("SIGINT", () => {
      readline.close();
      reject(new Error("认证已取消。"));
    });
  });
}

async function main(): Promise<void> {
  const [command, ...rest] = process.argv.slice(2);
  if (command === "--help" || command === "-h" || !command) {
    usage();
    process.exit(command ? 0 : 1);
  }
	if (command === "auth" && (rest.length === 0 || (rest.length === 1 && rest[0] === "--from-env"))) {
		const accessToken = rest.length === 0
			? await readSecret("accessToken（输入后不会回显）: ")
			: readAccessTokenFromEnv();
		await saveAccessToken(accessToken);
		console.log(`Unsplash 认证已保存到 ${AUTH_FILE}。该文件已被 Git 忽略。`);
    return;
  }
  if (command === "fetch") {
    await hydrateUnsplashAssets(["--external-agent", ...rest]);
    return;
  }
  if (command !== "auth" || rest.length) {
    usage();
    process.exit(1);
  }
}

main().catch((error) => {
  console.error(`Unsplash 认证失败: ${error instanceof Error ? error.message : String(error)}`);
  process.exit(1);
});
