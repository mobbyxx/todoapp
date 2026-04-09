import React, { useEffect, useCallback } from 'react';
import { View, StyleSheet, Dimensions } from 'react-native';
import Animated, {
  useSharedValue,
  useAnimatedStyle,
  withTiming,
  withDelay,
  Easing,
  runOnJS,
  interpolate,
} from 'react-native-reanimated';

const { width: SCREEN_WIDTH, height: SCREEN_HEIGHT } = Dimensions.get('window');

interface Particle {
  id: number;
  x: number;
  y: number;
  color: string;
  size: number;
  rotation: number;
  velocityX: number;
  velocityY: number;
  delay: number;
}

interface ConfettiProps {
  active: boolean;
  particleCount?: number;
  colors?: string[];
  duration?: number;
  onComplete?: () => void;
  origin?: { x: number; y: number };
}

const DEFAULT_COLORS = [
  '#FF6B6B',
  '#4ECDC4',
  '#45B7D1',
  '#FFA07A',
  '#98D8C8',
  '#F7DC6F',
  '#BB8FCE',
  '#85C1E2',
  '#F8B739',
  '#6C5CE7',
];

function createParticles(
  count: number,
  colors: string[],
  origin: { x: number; y: number }
): Particle[] {
  return Array.from({ length: count }, (_, i) => ({
    id: i,
    x: origin.x,
    y: origin.y,
    color: colors[Math.floor(Math.random() * colors.length)],
    size: 8 + Math.random() * 8,
    rotation: Math.random() * 360,
    velocityX: (Math.random() - 0.5) * 400,
    velocityY: -200 - Math.random() * 300,
    delay: Math.random() * 200,
  }));
}

interface ConfettiParticleProps {
  particle: Particle;
  duration: number;
  onParticleComplete: () => void;
}

function ConfettiParticle({
  particle,
  duration,
  onParticleComplete,
}: ConfettiParticleProps) {
  const progress = useSharedValue(0);
  const rotation = useSharedValue(particle.rotation);

  useEffect(() => {
    const animationDuration = duration;

    progress.value = withDelay(
      particle.delay,
      withTiming(1, {
        duration: animationDuration,
        easing: Easing.out(Easing.quad),
      }, () => {
        runOnJS(onParticleComplete)();
      })
    );

    rotation.value = withDelay(
      particle.delay,
      withTiming(particle.rotation + 720, {
        duration: animationDuration,
        easing: Easing.linear,
      })
    );
  }, []);

  const animatedStyle = useAnimatedStyle(() => {
    const translateX = particle.velocityX * progress.value;
    const translateY =
      particle.velocityY * progress.value +
      0.5 * 500 * progress.value * progress.value;

    const opacity = interpolate(
      progress.value,
      [0, 0.7, 1],
      [1, 1, 0]
    );

    const scale = interpolate(
      progress.value,
      [0, 0.1, 0.9, 1],
      [0, 1, 1, 0.8]
    );

    return {
      transform: [
        { translateX },
        { translateY },
        { rotate: `${rotation.value}deg` },
        { scale },
      ],
      opacity,
    };
  });

  return (
    <Animated.View
      style={[
        styles.particle,
        {
          width: particle.size,
          height: particle.size * 0.6,
          backgroundColor: particle.color,
          left: particle.x,
          top: particle.y,
        },
        animatedStyle,
      ]}
    />
  );
}

export function Confetti({
  active,
  particleCount = 50,
  colors = DEFAULT_COLORS,
  duration = 3000,
  onComplete,
  origin = { x: SCREEN_WIDTH / 2, y: SCREEN_HEIGHT / 3 },
}: ConfettiProps) {
  const [particles, setParticles] = React.useState<Particle[]>([]);
  const completedParticles = React.useRef(0);

  const handleParticleComplete = useCallback(() => {
    completedParticles.current += 1;
    if (completedParticles.current >= particleCount && onComplete) {
      onComplete();
    }
  }, [particleCount, onComplete]);

  useEffect(() => {
    if (active) {
      completedParticles.current = 0;
      setParticles(createParticles(particleCount, colors, origin));
    } else {
      setParticles([]);
    }
  }, [active, particleCount, colors, origin]);

  if (!active || particles.length === 0) {
    return null;
  }

  return (
    <View style={styles.container} pointerEvents="none">
      {particles.map((particle) => (
        <ConfettiParticle
          key={particle.id}
          particle={particle}
          duration={duration}
          onParticleComplete={handleParticleComplete}
        />
      ))}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    ...StyleSheet.absoluteFillObject,
    zIndex: 1000,
  },
  particle: {
    position: 'absolute',
    borderRadius: 2,
  },
});
