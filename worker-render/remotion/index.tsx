import { Composition, registerRoot } from "remotion";
import { UGCVertical } from "./UGCVertical";

registerRoot(() => (
  <>
    <Composition
      id="UGCVertical"
      component={UGCVertical}
      durationInFrames={600}
      fps={30}
      width={1080}
      height={1920}
      defaultProps={{ manifest: { scenes: [], captions: { words: [] } } }}
      calculateMetadata={({ props }) => {
        const manifest = props.manifest ?? {};
        const fps = manifest.format?.fps ?? 30;
        const introMs = manifest.introDurationMs ?? (manifest.introClip ? 2500 : 0);
        const scenesEnd =
          manifest.scenes?.reduce(
            (max, s) => Math.max(max, (s.startMs ?? 0) + (s.durationMs ?? 0)),
            0
          ) ?? 0;
        const totalMs = Math.max(introMs, scenesEnd, 20000);
        return {
          durationInFrames: Math.max(Math.ceil((totalMs / 1000) * fps), fps * 5),
          fps,
          width: manifest.format?.width ?? 1080,
          height: manifest.format?.height ?? 1920,
        };
      }}
    />
  </>
));
