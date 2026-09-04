import assert from "node:assert/strict";
import { mkdtemp, readFile, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import test from "node:test";

import { hydrateUnsplashAssets } from "./hydrate_unsplash_assets.ts";

test("fetch reuses one background asset for every page with the same content_type", async () => {
  const workDir = await mkdtemp(join(tmpdir(), "ppt-deck-planner-unsplash-"));
  const originalFetch = globalThis.fetch;
  const originalKey = process.env.UNSPLASH_ACCESS_KEY;
  let searchCalls = 0;
  let imageCalls = 0;
  try {
    await writeFile(join(workDir, "tasks.json"), JSON.stringify({
      tasks: [
        task("slide-1", "content_slide", "city skyline"),
        task("slide-2", "content_slide", "office meeting"),
        task("slide-3", "chart_slide", "factory"),
      ],
    }));
    process.env.UNSPLASH_ACCESS_KEY = "test-key";
    globalThis.fetch = async (input: string | URL | Request) => {
      const url = String(input);
      if (url.includes("/search/photos")) {
        searchCalls += 1;
        return jsonResponse({ results: [photo(`photo-${searchCalls}`)] });
      }
      if (url.includes("/download")) return jsonResponse({});
      imageCalls += 1;
      return imageResponse();
    };

    await hydrateUnsplashAssets(["--external-agent", "--work-dir", workDir]);

    const manifest = JSON.parse(await readFile(join(workDir, "tasks.json"), "utf8"));
    const [first, second, third] = manifest.tasks.map((taskValue: { content_plan: { visual_intent: Record<string, string> } }) => taskValue.content_plan.visual_intent);
    assert.equal(searchCalls, 2, "one search per content_type background group");
    assert.equal(imageCalls, 2, "one download per content_type background group");
    assert.equal(first.local_path, second.local_path);
    assert.equal(first.source_url, second.source_url);
    assert.equal(first.asset_query, second.asset_query, "the shared result also normalizes the background query");
    assert.notEqual(first.local_path, third.local_path);
  } finally {
    globalThis.fetch = originalFetch;
    if (originalKey === undefined) delete process.env.UNSPLASH_ACCESS_KEY;
    else process.env.UNSPLASH_ACCESS_KEY = originalKey;
    await rm(workDir, { recursive: true, force: true });
  }
});

function task(taskID: string, contentType: string, query: string) {
  return {
    task_id: taskID,
    content_type: contentType,
    content_plan: {
      visual_intent: {
        asset_purpose: "background",
        asset_query: query,
        asset_subject: query,
        composition: "wide landscape",
        orientation: "landscape",
      },
    },
  };
}

function photo(id: string) {
  return {
    id,
    width: 1600,
    height: 900,
    urls: {
      regular: `https://images.unsplash.com/${id}.jpg`,
      small: `https://images.unsplash.com/${id}-small.jpg`,
    },
    links: {
      html: `https://unsplash.com/photos/${id}`,
      download_location: `https://api.unsplash.com/photos/${id}/download`,
    },
    user: { name: "Test Photographer", links: { html: "https://unsplash.com/@test" } },
  };
}

function jsonResponse(value: unknown): Response {
  return { ok: true, json: async () => value } as Response;
}

function imageResponse(): Response {
  return {
    ok: true,
    url: "https://images.unsplash.com/test.jpg",
    headers: new Headers({ "content-type": "image/jpeg" }),
    arrayBuffer: async () => new Uint8Array([0xff, 0xd8, 0xff, 0xd9]).buffer,
  } as Response;
}
