import React from "react";
import {
  AbsoluteFill,
  Audio,
  OffthreadVideo,
  interpolate,
  useCurrentFrame,
  useVideoConfig,
} from "remotion";

type Manifest = {
  scenes?: Array<{
    id: string;
    startMs: number;
    durationMs: number;
    background: string;
    narration: string;
    videoUrl?: string;
  }>;
  captions?: {
    style?: string;
    words?: Array<{ text: string; startMs: number; endMs: number }>;
  };
  audio?: { narrationUrl?: string; musicVolume?: number };
  productName?: string;
};

const mediaSrc = (url: string) =>
  url.startsWith("http") ? url : `http://localhost:8080${url}`;

export const UGCVertical: React.FC<{ manifest: Manifest }> = ({ manifest }) => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();
  const ms = (frame / fps) * 1000;

  const scene =
    manifest.scenes?.find((s) => ms >= s.startMs && ms < s.startMs + s.durationMs) ??
    manifest.scenes?.[0];

  const activeWords =
    manifest.captions?.words?.filter((w) => ms >= w.startMs && ms < w.endMs) ?? [];

  const bg = scene?.background ?? "#1a1a2e";

  return (
    <AbsoluteFill style={{ backgroundColor: bg, fontFamily: "system-ui, sans-serif" }}>
      {scene?.videoUrl ? (
        <AbsoluteFill>
          <OffthreadVideo
            src={mediaSrc(scene.videoUrl)}
            style={{ width: "100%", height: "100%", objectFit: "cover" }}
          />
          <AbsoluteFill style={{ backgroundColor: "rgba(0,0,0,0.35)" }} />
        </AbsoluteFill>
      ) : null}

      {manifest.audio?.narrationUrl ? (
        <Audio src={mediaSrc(manifest.audio.narrationUrl)} />
      ) : null}

      <AbsoluteFill
        style={{
          justifyContent: "flex-end",
          alignItems: "center",
          padding: "120px 48px",
        }}
      >
        <div
          style={{
            color: "white",
            fontSize: 48,
            fontWeight: 800,
            textAlign: "center",
            lineHeight: 1.2,
            textShadow: "0 4px 24px rgba(0,0,0,0.8)",
          }}
        >
          {activeWords.length > 0
            ? activeWords.map((w, i) => (
                <span
                  key={`${w.text}-${i}`}
                  style={{
                    color: i === activeWords.length - 1 ? "#FFE066" : "white",
                    marginRight: 12,
                  }}
                >
                  {w.text}
                </span>
              ))
            : scene?.narration ?? manifest.productName}
        </div>
      </AbsoluteFill>

      <SceneProgress ms={ms} manifest={manifest} />
    </AbsoluteFill>
  );
};

const SceneProgress: React.FC<{ ms: number; manifest: Manifest }> = ({ ms, manifest }) => {
  const scenes = manifest.scenes ?? [];
  if (scenes.length === 0) return null;
  const total = scenes.reduce((m, s) => Math.max(m, s.startMs + s.durationMs), 1);
  const progress = interpolate(ms, [0, total], [0, 100], { extrapolateRight: "clamp" });
  return (
    <div
      style={{
        position: "absolute",
        bottom: 0,
        left: 0,
        height: 6,
        width: `${progress}%`,
        background: "#FFE066",
      }}
    />
  );
};
