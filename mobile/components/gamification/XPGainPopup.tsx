import { View, Text, StyleSheet } from 'react-native';
import Animated, {
  useSharedValue,
  useAnimatedStyle,
  withSpring,
  withSequence,
  withDelay,
  withTiming,
  runOnJS,
} from 'react-native-reanimated';
import { useEffect } from 'react';

interface XPGainPopupProps {
  visible: boolean;
  amount: number;
  reason?: string;
  onClose?: () => void;
}

export function XPGainPopup({
  visible,
  amount,
  reason,
  onClose,
}: XPGainPopupProps) {
  const translateY = useSharedValue(100);
  const opacity = useSharedValue(0);
  const scale = useSharedValue(0.5);

  useEffect(() => {
    if (visible) {
      translateY.value = withSequence(
        withSpring(0, { damping: 12, stiffness: 200 }),
        withDelay(
          1500,
          withTiming(-50, { duration: 300 }, () => {
            runOnJS(onClose || (() => {}))();
          })
        )
      );
      opacity.value = withSequence(
        withTiming(1, { duration: 200 }),
        withDelay(1500, withTiming(0, { duration: 300 }))
      );
      scale.value = withSequence(
        withSpring(1.2, { damping: 10, stiffness: 200 }),
        withTiming(1, { duration: 200 }),
        withDelay(1300, withTiming(0.8, { duration: 300 }))
      );
    }
  }, [visible]);

  const animatedStyle = useAnimatedStyle(() => ({
    transform: [
      { translateY: translateY.value },
      { scale: scale.value },
    ],
    opacity: opacity.value,
  }));

  if (!visible) return null;

  return (
    <View style={styles.overlay} pointerEvents="none">
      <Animated.View style={[styles.container, animatedStyle]}>
        <View style={styles.content}>
          <Text style={styles.amount}>+{amount}</Text>
          <Text style={styles.label}>XP</Text>
        </View>
        {reason && <Text style={styles.reason}>{reason}</Text>}
      </Animated.View>
    </View>
  );
}

const styles = StyleSheet.create({
  overlay: {
    ...StyleSheet.absoluteFillObject,
    justifyContent: 'center',
    alignItems: 'center',
    zIndex: 1000,
    pointerEvents: 'none',
  },
  container: {
    backgroundColor: 'rgba(59, 130, 246, 0.95)',
    borderRadius: 20,
    paddingHorizontal: 32,
    paddingVertical: 24,
    alignItems: 'center',
    shadowColor: '#3b82f6',
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.4,
    shadowRadius: 12,
    elevation: 8,
  },
  content: {
    flexDirection: 'row',
    alignItems: 'baseline',
    gap: 4,
  },
  amount: {
    fontSize: 48,
    fontWeight: 'bold',
    color: '#fff',
  },
  label: {
    fontSize: 24,
    fontWeight: '600',
    color: 'rgba(255, 255, 255, 0.9)',
  },
  reason: {
    fontSize: 14,
    color: 'rgba(255, 255, 255, 0.8)',
    marginTop: 8,
  },
});
