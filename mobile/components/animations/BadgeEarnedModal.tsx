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
import { Badge } from '../../types';

const { width: SCREEN_WIDTH, height: SCREEN_HEIGHT } = Dimensions.get('window');

interface BadgeEarnedModalProps {
  visible: boolean;
  badge: Badge | null;
  onClose: () => void;
  duration?: number;
}

export function BadgeEarnedModal({
  visible,
  badge,
  onClose,
  duration = 3000,
}: BadgeEarnedModalProps) {
  const [showConfetti, setShowConfetti] = useState(false);
  const slideY = useSharedValue(SCREEN_HEIGHT);
  const modalOpacity = useSharedValue(0);
  const badgeScale = useSharedValue(0);
  const badgeRotation = useSharedValue(0);
  const textOpacity = useSharedValue(0);
  const pulseScale = useSharedValue(1);

  useEffect(() => {
    if (visible && badge) {
      setShowConfetti(true);

      slideY.value = SCREEN_HEIGHT;
      modalOpacity.value = 0;
      badgeScale.value = 0;
      badgeRotation.value = 0;
      textOpacity.value = 0;
      pulseScale.value = 1;

      const delay = 100;

      modalOpacity.value = withDelay(
        delay,
        withTiming(1, { duration: 300 })
      );

      slideY.value = withDelay(
        delay,
        withSpring(0, {
          damping: 15,
          stiffness: 100,
        })
      );

      badgeScale.value = withDelay(
        delay + 300,
        withSequence(
          withTiming(1.2, { duration: 300, easing: Easing.out(Easing.back(2)) }),
          withTiming(1, { duration: 200 })
        )
      );

      badgeRotation.value = withDelay(
        delay + 300,
        withTiming(360, { duration: 800, easing: Easing.out(Easing.quad) })
      );

      pulseScale.value = withDelay(
        delay + 800,
        withSequence(
          withTiming(1.1, { duration: 400, easing: Easing.inOut(Easing.ease) }),
          withTiming(1, { duration: 400, easing: Easing.inOut(Easing.ease) })
        )
      );

      textOpacity.value = withDelay(
        delay + 600,
        withTiming(1, { duration: 400 })
      );

      const timeout = setTimeout(() => {
        setShowConfetti(false);
      }, duration);

      return () => clearTimeout(timeout);
    }
    slideY.value = SCREEN_HEIGHT;
    modalOpacity.value = 0;
    setShowConfetti(false);
    return undefined;
  }, [visible, badge, duration]);

  const modalAnimatedStyle = useAnimatedStyle(() => {
    return {
      opacity: modalOpacity.value,
      transform: [{ translateY: slideY.value }],
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

  const pulseAnimatedStyle = useAnimatedStyle(() => {
    return {
      transform: [{ scale: pulseScale.value }],
    };
  });

  const textAnimatedStyle = useAnimatedStyle(() => {
    return {
      opacity: textOpacity.value,
      transform: [
        {
          translateY: interpolate(textOpacity.value, [0, 1], [30, 0]),
        },
      ],
    };
  });

  const getBadgeIcon = (type: string): string => {
    switch (type) {
      case 'achievement':
        return '🏆';
      case 'milestone':
        return '🎯';
      case 'special':
        return '✨';
      default:
        return '🏅';
    }
  };

  const getBadgeColor = (type: string): string => {
    switch (type) {
      case 'achievement':
        return '#FFD700';
      case 'milestone':
        return '#9B59B6';
      case 'special':
        return '#E74C3C';
      default:
        return '#3498DB';
    }
  };

  if (!visible || !badge) {
    return null;
  }

  const badgeColor = getBadgeColor(badge.type);
  const badgeIcon = getBadgeIcon(badge.type);

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
          particleCount={60}
          duration={2500}
          origin={{ x: SCREEN_WIDTH / 2, y: SCREEN_HEIGHT / 3 }}
        />

        <Animated.View style={[styles.modalContainer, modalAnimatedStyle]}>
          <View style={styles.content}>
            <View style={[styles.badgeGlow, { backgroundColor: `${badgeColor}40` }]} />

            <Animated.View style={[styles.badgeContainer, badgeAnimatedStyle]}>
              <Animated.View style={[styles.badgeInner, pulseAnimatedStyle]}>
                <View style={[styles.badge, { backgroundColor: badgeColor }]}>
                  <Text style={styles.badgeIcon}>{badgeIcon}</Text>
                </View>
              </Animated.View>
              <View style={styles.sparkles}>
                <Text style={styles.sparkleIcon}>✨</Text>
              </View>
            </Animated.View>

            <Animated.View style={textAnimatedStyle}>
              <View style={styles.newBadgeTag}>
                <Text style={styles.newBadgeText}>NEW BADGE!</Text>
              </View>

              <Text style={styles.title}>{badge.name}</Text>
              <Text style={styles.description}>{badge.description}</Text>

              <View style={[styles.typeTag, { backgroundColor: `${badgeColor}20` }]}>
                <Text style={[styles.typeText, { color: badgeColor }]}>
                  {badge.type.charAt(0).toUpperCase() + badge.type.slice(1)}
                </Text>
              </View>

              {badge.points_value > 0 && (
                <View style={styles.pointsContainer}>
                  <Text style={styles.pointsIcon}>💎</Text>
                  <Text style={styles.pointsText}>+{badge.points_value} Points</Text>
                </View>
              )}
            </Animated.View>

            <TouchableOpacity style={[styles.button, { backgroundColor: badgeColor }]} onPress={onClose}>
              <Text style={styles.buttonText}>Collect</Text>
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
    justifyContent: 'flex-end',
  },
  modalContainer: {
    width: SCREEN_WIDTH,
  },
  content: {
    backgroundColor: '#fff',
    borderTopLeftRadius: 32,
    borderTopRightRadius: 32,
    padding: 32,
    paddingBottom: 48,
    alignItems: 'center',
    shadowColor: '#000',
    shadowOffset: { width: 0, height: -10 },
    shadowOpacity: 0.2,
    shadowRadius: 20,
    elevation: 20,
  },
  badgeGlow: {
    position: 'absolute',
    width: 180,
    height: 180,
    borderRadius: 90,
    top: 40,
  },
  badgeContainer: {
    width: 120,
    height: 120,
    marginBottom: 24,
    position: 'relative',
  },
  badgeInner: {
    width: 120,
    height: 120,
  },
  badge: {
    width: 120,
    height: 120,
    borderRadius: 60,
    justifyContent: 'center',
    alignItems: 'center',
    borderWidth: 6,
    borderColor: '#fff',
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 6 },
    shadowOpacity: 0.3,
    shadowRadius: 12,
    elevation: 12,
  },
  badgeIcon: {
    fontSize: 56,
  },
  sparkles: {
    position: 'absolute',
    top: -8,
    right: -8,
    backgroundColor: '#fff',
    borderRadius: 16,
    padding: 6,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.2,
    shadowRadius: 4,
    elevation: 4,
  },
  sparkleIcon: {
    fontSize: 20,
  },
  newBadgeTag: {
    backgroundColor: '#FF6B6B',
    paddingHorizontal: 16,
    paddingVertical: 6,
    borderRadius: 16,
    marginBottom: 16,
  },
  newBadgeText: {
    color: '#fff',
    fontSize: 12,
    fontWeight: '900',
    letterSpacing: 1,
  },
  title: {
    fontSize: 28,
    fontWeight: '800',
    color: '#1a1a1a',
    marginBottom: 12,
    textAlign: 'center',
  },
  description: {
    fontSize: 16,
    color: '#666',
    textAlign: 'center',
    marginBottom: 16,
    lineHeight: 24,
    paddingHorizontal: 16,
  },
  typeTag: {
    paddingHorizontal: 16,
    paddingVertical: 6,
    borderRadius: 16,
    marginBottom: 16,
  },
  typeText: {
    fontSize: 14,
    fontWeight: '700',
  },
  pointsContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 6,
    marginBottom: 24,
  },
  pointsIcon: {
    fontSize: 18,
  },
  pointsText: {
    fontSize: 16,
    fontWeight: '700',
    color: '#9B59B6',
  },
  button: {
    paddingHorizontal: 56,
    paddingVertical: 16,
    borderRadius: 28,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.3,
    shadowRadius: 8,
    elevation: 6,
  },
  buttonText: {
    color: '#fff',
    fontSize: 18,
    fontWeight: '800',
  },
});
