import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));

/**
 * Phase 1 render entrypoint.
 * Usage: node scripts/render.mjs <manifest.json> <output.mp4>
 *
 * When Remotion deps are installed, this bundles UGCVertical composition.
 * Falls back to copying mock bytes when RENDER_MOCK=1 or bundle fails.
 */
async function main() {
  const [manifestPath, outputPath] = process.argv.slice(2);
  if (!manifestPath || !outputPath) {
    console.error("Usage: node render.mjs <manifest.json> <output.mp4>");
    process.exit(1);
  }

  const manifest = JSON.parse(fs.readFileSync(manifestPath, "utf8"));
  console.log("Rendering run", manifest.runId, "scenes:", manifest.scenes?.length ?? 0);

  if (process.env.RENDER_MOCK === "1" || process.env.RENDER_MOCK === "true") {
    writeMock(outputPath);
    return;
  }

  try {
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
  } catch (err) {
    console.warn("Remotion render failed, using fallback:", err.message);
    writeMock(outputPath);
  }
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
