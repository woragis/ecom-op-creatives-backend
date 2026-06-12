# worker-render (Node)

Remotion UGC 9:16 composition.

```bash
npm install
RENDER_MOCK=1 node scripts/render.mjs path/to/manifest.json path/to/output.mp4
```

Real render (requires Chromium, ~2GB RAM):

```bash
RENDER_MOCK=0 API_PUBLIC_URL=http://localhost:8080 node scripts/render.mjs manifest.json output.mp4
```

Go `worker-pipeline` invokes this when `RENDER_MOCK=0` in `.env`.

See [../../docs/RENDER.md](../../docs/RENDER.md).
