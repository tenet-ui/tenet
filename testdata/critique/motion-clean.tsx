import { AbsoluteFill, Easing, Sequence, interpolate, spring, useCurrentFrame } from "remotion";
import { accent, damping, fps, ink, mass, paper, settle, stagger, stiffness, ms, frames, family, weightDisplay } from "./tokens";

const easing = Easing.bezier(...settle);

const Title: React.FC<{ words: string[] }> = ({ words }) => {
  const frame = useCurrentFrame();
  return (
    <div style={{ color: ink, fontFamily: family, fontWeight: weightDisplay }}>
      {words.map((word, i) => {
        const start = Math.min(i, stagger.limit) * ms(stagger.step);
        const progress = interpolate(frame - start, [0, frames.base], [0, 1], {
          easing,
          extrapolateLeft: "clamp",
          extrapolateRight: "clamp",
        });
        return (
          <span key={word} style={{ opacity: progress, transform: `translateY(${(1 - progress) * 24}px)` }}>
            {word}{" "}
          </span>
        );
      })}
    </div>
  );
};

const Counter: React.FC<{ value: number }> = ({ value }) => {
  const frame = useCurrentFrame();
  const progress = spring({
    frame,
    fps,
    config: { damping, stiffness, mass, overshootClamping: true },
    durationInFrames: frames.hero,
  });
  return <div style={{ color: accent }}>{Math.round(progress * value)}</div>;
};

export const TitleAndProof: React.FC = () => (
  <AbsoluteFill style={{ backgroundColor: paper }}>
    <Sequence from={0} durationInFrames={frames.hero}>
      <Title words={["Settled", "by", "design"]} />
    </Sequence>
    <Sequence from={frames.hero} durationInFrames={frames.hero + fps}>
      <Counter value={98} />
    </Sequence>
  </AbsoluteFill>
);
