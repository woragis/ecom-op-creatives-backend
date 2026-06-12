# worker-llm (Phase 0 stub)

Phase 0 uses a single stub worker that consumes **all** pipeline queues and marks steps as done with placeholder output.

Entry point (same Go module as the API):

```bash
cd server && go run ./cmd/worker-stub
```

Or from backend root:

```bash
make worker-stub
```

Future phases will split into dedicated workers:

- `worker-llm` — research, hooks, script, director, prompter, supervisor
- `worker-media` — ElevenLabs, image/video APIs, FFmpeg
- `worker-render` — Remotion (Node)

See [../docs/WORKERS.md](../../docs/WORKERS.md).
