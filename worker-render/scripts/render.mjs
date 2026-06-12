import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));

/**
 * Remotion render entrypoint.
 * Usage: node scripts/render.mjs <manifest.json> <output.mp4>
 *
 * Set RENDER_MOCK=1 to write stub bytes (dev without Remotion deps).
 * With RENDER_MOCK=0, Remotion must be installed (npm install in worker-render).
 *
 * Env: CREATIVE_RUN_ID, RENDER_LOG_PATH (optional append-only JSONL log on disk).
 */
function logJson(level, msg, extra = {}) {
  const line = JSON.stringify({
    time: new Date().toISOString(),
    level,
    service: "remotion",
    run_id: process.env.CREATIVE_RUN_ID || null,
    msg,
    ...extra,
  });
  console.log(line);
  const logPath = process.env.RENDER_LOG_PATH;
  if (logPath) {
    fs.mkdirSync(path.dirname(logPath), { recursive: true });
    fs.appendFileSync(logPath, line + "\n");
  }
}

async function main() {
  const [manifestPath, outputPath] = process.argv.slice(2);
  if (!manifestPath || !outputPath) {
    console.error("Usage: node render.mjs <manifest.json> <output.mp4>");
    process.exit(1);
  }

  const manifest = JSON.parse(fs.readFileSync(manifestPath, "utf8"));
  const apiBase = process.env.API_PUBLIC_URL ?? "http://localhost:8080";
  const runId = process.env.CREATIVE_RUN_ID || manifest.runId || null;
  const started = Date.now();

  logJson("INFO", "render started", {
    run_id: runId,
    manifest_path: manifestPath,
    output_path: outputPath,
    scenes: manifest.scenes?.length ?? 0,
    intro_ms: manifest.introDurationMs ?? 0,
    media_base: apiBase,
  });

  if (process.env.RENDER_MOCK === "1" || process.env.RENDER_MOCK === "true") {
    writeMock(outputPath);
    logJson("INFO", "render completed (mock)", {
      run_id: runId,
      output_path: outputPath,
      duration_ms: Date.now() - started,
    });
    return;
  }

  const { bundle } = await import("@remotion/bundler");
  const { renderMedia, selectComposition } = await import("@remotion/renderer");

  const entry = path.join(__dirname, "..", "remotion", "index.tsx");
  logJson("DEBUG", "bundling composition", { run_id: runId, entry });
  const bundleLocation = await bundle({ entryPoint: entry });

  const composition = await selectComposition({
    serveUrl: bundleLocation,
    id: "UGCVertical",
    inputProps: { manifest },
  });

  logJson("INFO", "rendering media", {
    run_id: runId,
    composition_id: composition.id,
    duration_frames: composition.durationInFrames,
    fps: composition.fps,
  });

  await renderMedia({
    composition,
    serveUrl: bundleLocation,
    codec: "h264",
    outputLocation: outputPath,
    inputProps: { manifest },
  });

  logJson("INFO", "render completed", {
    run_id: runId,
    output_path: outputPath,
    duration_ms: Date.now() - started,
  });
}

function writeMock(outputPath) {
  fs.mkdirSync(path.dirname(outputPath), { recursive: true });
  fs.writeFileSync(outputPath, Buffer.from("MOCK_MP4_PHASE1"));
}

main().catch((err) => {
  logJson("ERROR", "render failed", {
    run_id: process.env.CREATIVE_RUN_ID || null,
    error: err?.message || String(err),
  });
  process.exit(1);
});
