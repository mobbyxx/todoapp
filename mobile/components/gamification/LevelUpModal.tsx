import { View, Text, StyleSheet, Modal, TouchableOpacity } from 'react-native';
import Animated, {
  useSharedValue,
  useAnimatedStyle,
  withSpring,
  withSequence,
  withTiming,
  withDelay,
  interpolate,
  runOnJS,
} from 'react-native-reanimated';
import { useEffect } from 'react';
import { Level } from '../../types';

interface LevelUpModalProps {
  visible: boolean;
  level: Level;
  onClose?: () => void;
}

export function LevelUpModal({ visible, level, onClose }: LevelUpModalProps) {
  const scale = useSharedValue(0);
  const opacity = useSharedValue(0);
  const rotation = useSharedValue(0);

  useEffect(() => {
    if (visible) {
      scale.value = withSequence(
        withSpring(1, { damping: 10, stiffness: 100 }),
        withDelay(2000, withTiming(0.9, { duration: 200 }))
      );
      opacity.value = withSequence(
        withTiming(1, { duration: 300 }),
        withDelay(2000, withTiming(0, { duration: 300 }, () => {
          runOnJS(onClose || (() => {}))();
        }))
      );
      rotation.value = withSequence(
        withTiming(-10, { duration: 100 }),
        withTiming(10, { duration: 100 }),
        withTiming(-10, { duration: 100 }),
        withTiming(10, { duration: 100 }),
        withTiming(0, { duration: 200 })
      );
    }
  }, [visible]);

  const containerStyle = useAnimatedStyle(() => ({
    transform: [{ scale: scale.value }],
    opacity: opacity.value,
  }));

  const badgeStyle = useAnimatedStyle(() => ({
    transform: [{ rotate: `${rotation.value}deg` }],
  }));

  const particles = Array.from({ length: 12 }, (_, i) => {
    const angle = (i / 12) * 360;
    const particleScale = useSharedValue(0);
    const particleOpacity = useSharedValue(0);

    useEffect(() => {
      if (visible) {
        particleScale.value = withSequence(
          withDelay(i * 50, withSpring(1, { damping: 15, stiffness: 200 })),
          withDelay(1500, withTiming(0, { duration: 300 }))
        );
        particleOpacity.value = withSequence(
          withDelay(i * 50, withTiming(1, { duration: 200 })),
          withDelay(1500, withTiming(0, { duration: 300 }))
        );
      }
    }, [visible]);

    const particleStyle = useAnimatedStyle(() => ({
      transform: [
        { rotate: `${angle}deg` },
        { translateY: interpolate(particleScale.value, [0, 1], [0, -80]) },
        { scale: particleScale.value },
      ],
      opacity: particleOpacity.value,
    }));

    return particleStyle;
  });

  if (!visible) return null;

  return (
    <Modal visible={visible} transparent animationType="none">
      <View style={styles.overlay}>
        <Animated.View style={[styles.container, containerStyle]}>
          {particles.map((style, index) => (
            <Animated.View
              key={index}
              style={[styles.particle, style]}
            >
              <Text style={styles.particleEmoji}>
                {['✨', '🎉', '⭐', '🌟', '💫'][index % 5]}
              </Text>
            </Animated.View>
          ))}

          <Animated.View style={[styles.badgeContainer, badgeStyle]}>
            <View style={styles.badge}>
              <Text style={styles.badgeIcon}>🏆</Text>
              <Text style={styles.levelNumber}>{level.level_number}</Text>
            </View>
          </Animated.View>

          <Text style={styles.title}>Level Up!</Text>
          <Text style={styles.levelName}>{level.name}</Text>
          <Text style={styles.description}>
            You've reached level {level.level_number}!
          </Text>

          <TouchableOpacity
            style={styles.closeButton}
            onPress={() => onClose?.()}
            activeOpacity={0.8}
          >
            <Text style={styles.closeButtonText}>Awesome!</Text>
          </TouchableOpacity>
        </Animated.View>
      </View>
    </Modal>
  );
}

const styles = StyleSheet.create({
  overlay: {
    flex: 1,
    backgroundColor: 'rgba(0, 0, 0, 0.7)',
    justifyContent: 'center',
    alignItems: 'center',
  },
  container: {
    backgroundColor: '#fff',
    borderRadius: 24,
    padding: 32,
    alignItems: 'center',
    width: '80%',
    maxWidth: 320,
    position: 'relative',
  },
  particle: {
    position: 'absolute',
    top: '50%',
    left: '50%',
    marginLeft: -12,
    marginTop: -12,
  },
  particleEmoji: {
    fontSize: 24,
  },
  badgeContainer: {
    marginBottom: 24,
  },
  badge: {
    width: 100,
    height: 100,
    borderRadius: 50,
    backgroundColor: '#fbbf24',
    alignItems: 'center',
    justifyContent: 'center',
    borderWidth: 4,
    borderColor: '#f59e0b',
    shadowColor: '#f59e0b',
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.5,
    shadowRadius: 8,
    elevation: 8,
  },
  badgeIcon: {
    fontSize: 40,
    marginBottom: -8,
  },
  levelNumber: {
    fontSize: 28,
    fontWeight: 'bold',
    color: '#fff',
  },
  title: {
    fontSize: 28,
    fontWeight: 'bold',
    color: '#1a1a1a',
    marginBottom: 8,
  },
  levelName: {
    fontSize: 20,
    fontWeight: '600',
    color: '#3b82f6',
    marginBottom: 8,
  },
  description: {
    fontSize: 14,
    color: '#666',
    textAlign: 'center',
    marginBottom: 24,
  },
  closeButton: {
    backgroundColor: '#3b82f6',
    paddingHorizontal: 32,
    paddingVertical: 12,
    borderRadius: 12,
  },
  closeButtonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '600',
  },
});
