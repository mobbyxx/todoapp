import React, { useEffect, useState } from 'react';
import {
  View,
  Text,
  StyleSheet,
  Modal,
  TouchableOpacity,
  Dimensions,
} from 'react-native';
import Animated, {
  useSharedValue,
  useAnimatedStyle,
  withTiming,
  withSequence,
  withDelay,
  withSpring,
  Easing,
  interpolate,
} from 'react-native-reanimated';
import { Confetti } from './Confetti';

const { width: SCREEN_WIDTH, height: SCREEN_HEIGHT } = Dimensions.get('window');

interface LevelUpModalProps {
  visible: boolean;
  levelNumber: number;
  levelName: string;
  previousLevel?: number;
  onClose: () => void;
  duration?: number;
}

export function LevelUpModal({
  visible,
  levelNumber,
  levelName,
  previousLevel,
  onClose,
  duration = 3000,
}: LevelUpModalProps) {
  const [showConfetti, setShowConfetti] = useState(false);
  const modalScale = useSharedValue(0);
  const modalOpacity = useSharedValue(0);
  const badgeRotation = useSharedValue(0);
  const badgeScale = useSharedValue(0);
  const textOpacity = useSharedValue(0);
  const glowOpacity = useSharedValue(0);

  useEffect(() => {
    if (visible) {
      setShowConfetti(true);

      modalScale.value = 0;
      modalOpacity.value = 0;
      badgeRotation.value = 0;
      badgeScale.value = 0;
      textOpacity.value = 0;
      glowOpacity.value = 0;

      const delay = 100;
      const badgeDuration = 600;

      modalOpacity.value = withDelay(
        delay,
        withTiming(1, { duration: 200 })
      );

      modalScale.value = withDelay(
        delay,
        withSpring(1, {
          damping: 12,
          stiffness: 100,
        })
      );

      badgeScale.value = withDelay(
        delay + 200,
        withSequence(
          withTiming(1.3, { duration: 200, easing: Easing.out(Easing.back(2)) }),
          withTiming(1, { duration: 150 })
        )
      );

      badgeRotation.value = withDelay(
        delay + 200,
        withTiming(360, { duration: badgeDuration, easing: Easing.out(Easing.quad) })
      );

      glowOpacity.value = withDelay(
        delay + 200,
        withSequence(
          withTiming(1, { duration: 300 }),
          withTiming(0.5, { duration: 500 })
        )
      );

      textOpacity.value = withDelay(
        delay + 400,
        withTiming(1, { duration: 400 })
      );

      const timeout = setTimeout(() => {
        setShowConfetti(false);
      }, duration);

      return () => clearTimeout(timeout);
    }
    modalScale.value = 0;
    modalOpacity.value = 0;
    setShowConfetti(false);
    return undefined;
  }, [visible, duration]);

  const modalAnimatedStyle = useAnimatedStyle(() => {
    return {
      opacity: modalOpacity.value,
      transform: [{ scale: modalScale.value }],
    };
  });

  const badgeAnimatedStyle = useAnimatedStyle(() => {
    return {
      transform: [
        { rotate: `${badgeRotation.value}deg` },
        { scale: badgeScale.value },
      ],
    };
  });

  const glowAnimatedStyle = useAnimatedStyle(() => {
    const scale = interpolate(glowOpacity.value, [0, 1], [0.8, 1.2]);
    return {
      opacity: glowOpacity.value,
      transform: [{ scale }],
    };
  });

  const textAnimatedStyle = useAnimatedStyle(() => {
    return {
      opacity: textOpacity.value,
      transform: [
        {
          translateY: interpolate(textOpacity.value, [0, 1], [20, 0]),
        },
      ],
    };
  });

  return (
    <Modal
      visible={visible}
      transparent
      animationType="none"
      onRequestClose={onClose}
      statusBarTranslucent
    >
      <View style={styles.overlay}>
        <Confetti
          active={showConfetti}
          particleCount={80}
          duration={2500}
          origin={{ x: SCREEN_WIDTH / 2, y: SCREEN_HEIGHT / 2 }}
        />

        <Animated.View style={[styles.modalContainer, modalAnimatedStyle]}>
          <View style={styles.content}>
            <Animated.View style={[styles.glowEffect, glowAnimatedStyle]} />

            <Animated.View style={[styles.badgeContainer, badgeAnimatedStyle]}>
              <View style={styles.badge}>
                <Text style={styles.levelNumber}>{levelNumber}</Text>
              </View>
              <View style={styles.starBurst}>
                <Text style={styles.starIcon}>⭐</Text>
              </View>
            </Animated.View>

            <Animated.View style={textAnimatedStyle}>
              <Text style={styles.title}>Level Up!</Text>
              <Text style={styles.levelName}>{levelName}</Text>
              {previousLevel && (
                <Text style={styles.previousLevel}>
                  {previousLevel} → {levelNumber}
                </Text>
              )}
              <Text style={styles.message}>
                Congratulations! You&apos;ve reached a new level!
              </Text>
            </Animated.View>

            <TouchableOpacity style={styles.button} onPress={onClose}>
              <Text style={styles.buttonText}>Awesome!</Text>
            </TouchableOpacity>
          </View>
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
  modalContainer: {
    width: SCREEN_WIDTH * 0.85,
    maxWidth: 340,
  },
  content: {
    backgroundColor: '#fff',
    borderRadius: 24,
    padding: 32,
    alignItems: 'center',
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 10 },
    shadowOpacity: 0.3,
    shadowRadius: 20,
    elevation: 20,
  },
  glowEffect: {
    position: 'absolute',
    width: 200,
    height: 200,
    borderRadius: 100,
    backgroundColor: 'rgba(255, 215, 0, 0.3)',
    top: 20,
  },
  badgeContainer: {
    width: 100,
    height: 100,
    marginBottom: 20,
    position: 'relative',
  },
  badge: {
    width: 100,
    height: 100,
    borderRadius: 50,
    backgroundColor: '#FFD700',
    justifyContent: 'center',
    alignItems: 'center',
    borderWidth: 4,
    borderColor: '#FFA500',
    shadowColor: '#FFD700',
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.5,
    shadowRadius: 10,
    elevation: 10,
  },
  levelNumber: {
    fontSize: 42,
    fontWeight: '900',
    color: '#B8860B',
  },
  starBurst: {
    position: 'absolute',
    top: -10,
    right: -10,
    width: 36,
    height: 36,
    borderRadius: 18,
    backgroundColor: '#fff',
    justifyContent: 'center',
    alignItems: 'center',
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.2,
    shadowRadius: 4,
    elevation: 4,
  },
  starIcon: {
    fontSize: 20,
  },
  title: {
    fontSize: 32,
    fontWeight: '900',
    color: '#1a1a1a',
    marginBottom: 8,
    textAlign: 'center',
  },
  levelName: {
    fontSize: 20,
    fontWeight: '700',
    color: '#007AFF',
    marginBottom: 4,
    textAlign: 'center',
  },
  previousLevel: {
    fontSize: 14,
    color: '#999',
    marginBottom: 16,
  },
  message: {
    fontSize: 16,
    color: '#666',
    textAlign: 'center',
    marginBottom: 24,
    lineHeight: 22,
  },
  button: {
    backgroundColor: '#007AFF',
    paddingHorizontal: 40,
    paddingVertical: 14,
    borderRadius: 25,
    shadowColor: '#007AFF',
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.3,
    shadowRadius: 8,
    elevation: 5,
  },
  buttonText: {
    color: '#fff',
    fontSize: 18,
    fontWeight: '700',
  },
});
