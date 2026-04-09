import React, { useEffect } from 'react';
import { View, Text, StyleSheet, Dimensions } from 'react-native';
import Animated, {
  useSharedValue,
  useAnimatedStyle,
  withTiming,
  withSequence,
  withDelay,
  Easing,
  interpolate,
} from 'react-native-reanimated';

const { width: SCREEN_WIDTH } = Dimensions.get('window');

interface XPGainPopupProps {
  visible: boolean;
  amount: number;
  reason?: string;
  position?: { x: number; y: number };
  onComplete?: () => void;
  duration?: number;
}

export function XPGainPopup({
  visible,
  amount,
  reason = '',
  position,
  onComplete,
  duration = 1500,
}: XPGainPopupProps) {
  const opacity = useSharedValue(0);
  const translateY = useSharedValue(0);
  const scale = useSharedValue(0.5);
  const shimmer = useSharedValue(0);

  const targetPosition = position || {
    x: SCREEN_WIDTH / 2,
    y: 150,
  };

  useEffect(() => {
    if (visible) {
      opacity.value = 0;
      translateY.value = 0;
      scale.value = 0.5;
      shimmer.value = 0;

      const animationDuration = duration;
      const fadeInDuration = animationDuration * 0.15;
      const floatDuration = animationDuration * 0.6;
      const fadeOutDuration = animationDuration * 0.25;

      opacity.value = withSequence(
        withTiming(1, { duration: fadeInDuration, easing: Easing.out(Easing.quad) }),
        withDelay(
          floatDuration,
          withTiming(0, { duration: fadeOutDuration, easing: Easing.in(Easing.quad) })
        )
      );

      scale.value = withSequence(
        withTiming(1.2, { duration: fadeInDuration, easing: Easing.out(Easing.back(2)) }),
        withTiming(1, { duration: 100 })
      );

      translateY.value = withSequence(
        withTiming(-80, {
          duration: animationDuration,
          easing: Easing.out(Easing.quad),
        })
      );

      shimmer.value = withTiming(1, { duration: animationDuration });

      const timeout = setTimeout(() => {
        if (onComplete) {
          onComplete();
        }
      }, animationDuration);

      return () => clearTimeout(timeout);
    }
    return undefined;
  }, [visible, duration]);

  const animatedContainerStyle = useAnimatedStyle(() => {
    return {
      opacity: opacity.value,
      transform: [
        { translateY: translateY.value },
        { scale: scale.value },
      ],
    };
  });

  const shimmerStyle = useAnimatedStyle(() => {
    const shimmerOpacity = interpolate(
      shimmer.value,
      [0, 0.3, 0.6, 1],
      [0.3, 1, 0.3, 0]
    );
    return {
      opacity: shimmerOpacity,
    };
  });

  if (!visible) {
    return null;
  }

  return (
    <View
      style={[
        styles.container,
        { left: targetPosition.x, top: targetPosition.y },
      ]}
      pointerEvents="none"
    >
      <Animated.View style={[styles.popup, animatedContainerStyle]}>
        <View style={styles.content}>
          <View style={styles.xpContainer}>
            <Text style={styles.xpIcon}>✨</Text>
            <Text style={styles.xpText}>+{amount} XP</Text>
          </View>
          {reason && <Text style={styles.reasonText}>{reason}</Text>}
        </View>
        <Animated.View style={[styles.shimmer, shimmerStyle]} />
      </Animated.View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    position: 'absolute',
    zIndex: 999,
    transform: [{ translateX: -75 }],
  },
  popup: {
    backgroundColor: '#007AFF',
    borderRadius: 25,
    paddingHorizontal: 20,
    paddingVertical: 12,
    minWidth: 150,
    shadowColor: '#007AFF',
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.3,
    shadowRadius: 8,
    elevation: 8,
    overflow: 'hidden',
  },
  content: {
    alignItems: 'center',
  },
  xpContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 6,
  },
  xpIcon: {
    fontSize: 20,
  },
  xpText: {
    fontSize: 20,
    fontWeight: '800',
    color: '#fff',
  },
  reasonText: {
    fontSize: 12,
    color: 'rgba(255, 255, 255, 0.9)',
    marginTop: 4,
    fontWeight: '500',
  },
  shimmer: {
    ...StyleSheet.absoluteFillObject,
    backgroundColor: 'rgba(255, 255, 255, 0.3)',
  },
});
