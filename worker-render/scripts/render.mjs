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
 */
async function main() {
  const [manifestPath, outputPath] = process.argv.slice(2);
  if (!manifestPath || !outputPath) {
    console.error("Usage: node render.mjs <manifest.json> <output.mp4>");
    process.exit(1);
  }

  const manifest = JSON.parse(fs.readFileSync(manifestPath, "utf8"));
  const apiBase = process.env.API_PUBLIC_URL ?? "http://localhost:8080";
  console.log(
    "Rendering run",
    manifest.runId,
    "scenes:",
    manifest.scenes?.length ?? 0,
    "introMs:",
    manifest.introDurationMs ?? 0,
    "media:",
    apiBase
  );

  if (process.env.RENDER_MOCK === "1" || process.env.RENDER_MOCK === "true") {
    writeMock(outputPath);
    return;
  }

  const { bundle } = await import("@remotion/bundler");
  const { renderMedia, selectComposition } = await import("@remotion/renderer");

  const entry = path.join(__dirname, "..", "remotion", "index.tsx");
  const bundleLocation = await bundle({ entryPoint: entry });

  const composition = await selectComposition({
    serveUrl: bundleLocation,
    id: "UGCVertical",
    inputProps: { manifest },
  });

  await renderMedia({
    composition,
    serveUrl: bundleLocation,
    codec: "h264",
    outputLocation: outputPath,
    inputProps: { manifest },
  });

  console.log("Rendered", outputPath);
}

function writeMock(outputPath) {
  fs.mkdirSync(path.dirname(outputPath), { recursive: true });
  fs.writeFileSync(outputPath, Buffer.from("MOCK_MP4_PHASE1"));
  console.log("Wrote mock output", outputPath);
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
