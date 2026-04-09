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

interface GoalCompletedModalProps {
  visible: boolean;
  goalTitle: string;
  goalDescription?: string;
  xpReward?: number;
  completedBy?: string;
  isSharedGoal?: boolean;
  onClose: () => void;
  onShare?: () => void;
  duration?: number;
}

export function GoalCompletedModal({
  visible,
  goalTitle,
  goalDescription,
  xpReward = 50,
  completedBy,
  isSharedGoal = false,
  onClose,
  onShare,
  duration = 3000,
}: GoalCompletedModalProps) {
  const [showConfetti, setShowConfetti] = useState(false);
  const modalScale = useSharedValue(0.3);
  const modalOpacity = useSharedValue(0);
  const trophyScale = useSharedValue(0);
  const trophyRotation = useSharedValue(0);
  const textOpacity = useSharedValue(0);
  const textReveal = useSharedValue(0);
  const checkmarkScale = useSharedValue(0);

  useEffect(() => {
    if (visible) {
      setShowConfetti(true);

      modalScale.value = 0.3;
      modalOpacity.value = 0;
      trophyScale.value = 0;
      trophyRotation.value = -180;
      textOpacity.value = 0;
      textReveal.value = 0;
      checkmarkScale.value = 0;

      const delay = 100;

      modalOpacity.value = withDelay(
        delay,
        withTiming(1, { duration: 300 })
      );

      modalScale.value = withDelay(
        delay,
        withSpring(1, {
          damping: 12,
          stiffness: 100,
        })
      );

      trophyScale.value = withDelay(
        delay + 200,
        withSequence(
          withTiming(1.3, { duration: 400, easing: Easing.out(Easing.back(1.7)) }),
          withTiming(1, { duration: 200 })
        )
      );

      trophyRotation.value = withDelay(
        delay + 200,
        withTiming(0, { duration: 600, easing: Easing.out(Easing.back(1.2)) })
      );

      checkmarkScale.value = withDelay(
        delay + 500,
        withTiming(1, { duration: 300, easing: Easing.out(Easing.back(2)) })
      );

      textOpacity.value = withDelay(
        delay + 400,
        withTiming(1, { duration: 500 })
      );

      textReveal.value = withDelay(
        delay + 500,
        withTiming(1, { duration: 600, easing: Easing.out(Easing.quad) })
      );

      const timeout = setTimeout(() => {
        setShowConfetti(false);
      }, duration);

      return () => clearTimeout(timeout);
    }
    modalScale.value = 0.3;
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

  const trophyAnimatedStyle = useAnimatedStyle(() => {
    return {
      transform: [
        { rotate: `${trophyRotation.value}deg` },
        { scale: trophyScale.value },
      ],
    };
  });

  const checkmarkAnimatedStyle = useAnimatedStyle(() => {
    return {
      transform: [{ scale: checkmarkScale.value }],
      opacity: checkmarkScale.value,
    };
  });

  const textAnimatedStyle = useAnimatedStyle(() => {
    return {
      opacity: textOpacity.value,
    };
  });

  const titleAnimatedStyle = useAnimatedStyle(() => {
    return {
      opacity: textOpacity.value,
      transform: [
        {
          translateY: interpolate(textReveal.value, [0, 1], [20, 0]),
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
          particleCount={70}
          duration={2800}
          colors={['#22c55e', '#4ade80', '#86efac', '#bbf7d0', '#fcd34d', '#fbbf24']}
          origin={{ x: SCREEN_WIDTH / 2, y: SCREEN_HEIGHT / 2 }}
        />

        <Animated.View style={[styles.modalContainer, modalAnimatedStyle]}>
          <View style={styles.content}>
            <View style={styles.iconContainer}>
              <Animated.View style={[styles.trophyContainer, trophyAnimatedStyle]}>
                <View style={styles.trophy}>
                  <Text style={styles.trophyIcon}>🏆</Text>
                </View>
                <Animated.View style={[styles.checkmarkContainer, checkmarkAnimatedStyle]}>
                  <View style={styles.checkmark}>
                    <Text style={styles.checkmarkIcon}>✓</Text>
                  </View>
                </Animated.View>
              </Animated.View>
            </View>

            <Animated.View style={textAnimatedStyle}>
              <View style={styles.achievementTag}>
                <Text style={styles.achievementText}>GOAL ACHIEVED!</Text>
              </View>
            </Animated.View>

            <Animated.View style={titleAnimatedStyle}>
              <Text style={styles.title}>{goalTitle}</Text>
              {goalDescription && (
                <Text style={styles.description}>{goalDescription}</Text>
              )}
            </Animated.View>

            <Animated.View style={textAnimatedStyle}>
              {isSharedGoal && completedBy && (
                <View style={styles.completedByContainer}>
                  <Text style={styles.completedByLabel}>Completed by</Text>
                  <Text style={styles.completedByName}>{completedBy}</Text>
                </View>
              )}

              <View style={styles.rewardContainer}>
                <View style={styles.xpBadge}>
                  <Text style={styles.xpIcon}>⚡</Text>
                  <Text style={styles.xpText}>+{xpReward} XP</Text>
                </View>
                {isSharedGoal && (
                  <View style={styles.sharedBadge}>
                    <Text style={styles.sharedIcon}>👥</Text>
                    <Text style={styles.sharedText}>Shared</Text>
                  </View>
                )}
              </View>

              <View style={styles.buttonContainer}>
                {onShare && (
                  <TouchableOpacity style={styles.shareButton} onPress={onShare}>
                    <Text style={styles.shareButtonText}>Share</Text>
                  </TouchableOpacity>
                )}
                <TouchableOpacity style={styles.closeButton} onPress={onClose}>
                  <Text style={styles.closeButtonText}>Awesome!</Text>
                </TouchableOpacity>
              </View>
            </Animated.View>
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
    width: SCREEN_WIDTH * 0.9,
    maxWidth: 380,
  },
  content: {
    backgroundColor: '#fff',
    borderRadius: 28,
    padding: 32,
    alignItems: 'center',
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 12 },
    shadowOpacity: 0.25,
    shadowRadius: 24,
    elevation: 24,
    borderWidth: 2,
    borderColor: '#22c55e',
  },
  iconContainer: {
    marginBottom: 20,
  },
  trophyContainer: {
    position: 'relative',
    width: 100,
    height: 100,
  },
  trophy: {
    width: 100,
    height: 100,
    borderRadius: 50,
    backgroundColor: '#22c55e',
    justifyContent: 'center',
    alignItems: 'center',
    shadowColor: '#22c55e',
    shadowOffset: { width: 0, height: 6 },
    shadowOpacity: 0.4,
    shadowRadius: 12,
    elevation: 12,
  },
  trophyIcon: {
    fontSize: 52,
  },
  checkmarkContainer: {
    position: 'absolute',
    bottom: -4,
    right: -4,
  },
  checkmark: {
    width: 36,
    height: 36,
    borderRadius: 18,
    backgroundColor: '#fff',
    justifyContent: 'center',
    alignItems: 'center',
    borderWidth: 3,
    borderColor: '#22c55e',
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.2,
    shadowRadius: 4,
    elevation: 4,
  },
  checkmarkIcon: {
    fontSize: 18,
    color: '#22c55e',
    fontWeight: '900',
  },
  achievementTag: {
    backgroundColor: '#22c55e',
    paddingHorizontal: 16,
    paddingVertical: 6,
    borderRadius: 16,
    marginBottom: 16,
  },
  achievementText: {
    color: '#fff',
    fontSize: 12,
    fontWeight: '900',
    letterSpacing: 1.5,
  },
  title: {
    fontSize: 26,
    fontWeight: '800',
    color: '#1a1a1a',
    marginBottom: 8,
    textAlign: 'center',
  },
  description: {
    fontSize: 15,
    color: '#666',
    textAlign: 'center',
    marginBottom: 20,
    lineHeight: 22,
    paddingHorizontal: 8,
  },
  completedByContainer: {
    alignItems: 'center',
    marginBottom: 16,
    paddingVertical: 12,
    paddingHorizontal: 24,
    backgroundColor: '#f3f4f6',
    borderRadius: 12,
  },
  completedByLabel: {
    fontSize: 12,
    color: '#9ca3af',
    marginBottom: 4,
    textTransform: 'uppercase',
    letterSpacing: 0.5,
  },
  completedByName: {
    fontSize: 16,
    fontWeight: '700',
    color: '#374151',
  },
  rewardContainer: {
    flexDirection: 'row',
    gap: 12,
    marginBottom: 24,
  },
  xpBadge: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 6,
    backgroundColor: '#fef3c7',
    paddingHorizontal: 16,
    paddingVertical: 10,
    borderRadius: 20,
    borderWidth: 2,
    borderColor: '#fcd34d',
  },
  xpIcon: {
    fontSize: 18,
  },
  xpText: {
    fontSize: 16,
    fontWeight: '800',
    color: '#d97706',
  },
  sharedBadge: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 6,
    backgroundColor: '#dbeafe',
    paddingHorizontal: 16,
    paddingVertical: 10,
    borderRadius: 20,
    borderWidth: 2,
    borderColor: '#93c5fd',
  },
  sharedIcon: {
    fontSize: 16,
  },
  sharedText: {
    fontSize: 14,
    fontWeight: '700',
    color: '#2563eb',
  },
  buttonContainer: {
    flexDirection: 'row',
    gap: 12,
    width: '100%',
  },
  shareButton: {
    flex: 1,
    backgroundColor: '#f3f4f6',
    paddingVertical: 14,
    borderRadius: 24,
    alignItems: 'center',
  },
  shareButtonText: {
    fontSize: 16,
    fontWeight: '700',
    color: '#374151',
  },
  closeButton: {
    flex: 1.5,
    backgroundColor: '#22c55e',
    paddingVertical: 14,
    borderRadius: 24,
    alignItems: 'center',
    shadowColor: '#22c55e',
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.3,
    shadowRadius: 8,
    elevation: 5,
  },
  closeButtonText: {
    fontSize: 16,
    fontWeight: '800',
    color: '#fff',
  },
});
