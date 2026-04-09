import { View, StyleSheet } from 'react-native';
import Animated, {
  useSharedValue,
  useAnimatedStyle,
  withTiming,
  withSpring,
} from 'react-native-reanimated';
import { useEffect } from 'react';

interface ProgressBarProps {
  progress: number;
  height?: number;
  animated?: boolean;
  showGlow?: boolean;
}

export function ProgressBar({
  progress,
  height = 8,
  animated = true,
  showGlow = false,
}: ProgressBarProps) {
  const progressValue = useSharedValue(0);
  const glowOpacity = useSharedValue(0);

  useEffect(() => {
    progressValue.value = withSpring(Math.max(0, Math.min(100, progress)), {
      damping: 15,
      stiffness: 100,
    });

    if (showGlow && progress >= 90) {
      glowOpacity.value = withTiming(1, { duration: 300 });
    } else {
      glowOpacity.value = withTiming(0, { duration: 200 });
    }
  }, [progress, showGlow]);

  const fillStyle = useAnimatedStyle(() => ({
    width: `${progressValue.value}%`,
  }));

  const glowStyle = useAnimatedStyle(() => ({
    opacity: glowOpacity.value,
  }));

  return (
    <View style={[styles.container, { height, borderRadius: height / 2 }]}>
      <View style={[styles.background, { borderRadius: height / 2 }]} />
      <Animated.View
        style={[
          styles.fill,
          { borderRadius: height / 2 },
          fillStyle,
          animated && styles.animatedFill,
        ]}
      />
      {showGlow && (
        <Animated.View
          style={[
            styles.glow,
            { borderRadius: height / 2 },
            glowStyle,
          ]}
          pointerEvents="none"
        />
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    width: '100%',
    overflow: 'hidden',
    position: 'relative',
  },
  background: {
    ...StyleSheet.absoluteFillObject,
    backgroundColor: '#e5e7eb',
  },
  fill: {
    height: '100%',
    backgroundColor: '#3b82f6',
  },
  animatedFill: {
    shadowColor: '#3b82f6',
    shadowOffset: { width: 0, height: 0 },
    shadowOpacity: 0.5,
    shadowRadius: 4,
    elevation: 2,
  },
  glow: {
    ...StyleSheet.absoluteFillObject,
    backgroundColor: 'rgba(59, 130, 246, 0.3)',
  },
});
