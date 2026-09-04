#!/usr/bin/env node
import { spawnSync } from "node:child_process";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const entry = resolve(dirname(fileURLToPath(import.meta.url)), "unsplash.ts");
const child = spawnSync(
  process.execPath,
  ["--no-warnings", "--experimental-strip-types", entry, ...process.argv.slice(2)],
  { stdio: "inherit" },
);

if (child.error) {
  console.error(`无法启动 Unsplash CLI: ${child.error.message}`);
  process.exit(1);
}
process.exit(child.status ?? 1);
