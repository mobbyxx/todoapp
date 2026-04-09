import { View, Text, StyleSheet, TouchableOpacity } from 'react-native';
import { Badge, UserBadge } from '../../types';

interface BadgeCardProps {
  badge: Badge;
  userBadge?: UserBadge;
  onPress?: (badge: Badge) => void;
  size?: 'small' | 'medium' | 'large';
}

const sizeConfig = {
  small: {
    container: 80,
    icon: 40,
    fontSize: 10,
    padding: 8,
  },
  medium: {
    container: 100,
    icon: 50,
    fontSize: 12,
    padding: 12,
  },
  large: {
    container: 120,
    icon: 60,
    fontSize: 14,
    padding: 16,
  },
};

export function BadgeCard({
  badge,
  userBadge,
  onPress,
  size = 'medium',
}: BadgeCardProps) {
  const isLocked = !userBadge;
  const config = sizeConfig[size];

  return (
    <TouchableOpacity
      style={[
        styles.container,
        {
          width: config.container,
          padding: config.padding,
          opacity: isLocked ? 0.5 : 1,
        },
      ]}
      onPress={() => onPress?.(badge)}
      activeOpacity={0.7}
      disabled={!onPress}
    >
      <View
        style={[
          styles.iconContainer,
          {
            width: config.icon,
            height: config.icon,
            borderRadius: config.icon / 2,
            backgroundColor: isLocked ? '#e5e7eb' : '#dbeafe',
          },
        ]}
      >
        <Text style={[styles.icon, { fontSize: config.icon * 0.5 }]}>
          {isLocked ? '🔒' : '🏆'}
        </Text>
      </View>
      <Text
        style={[styles.name, { fontSize: config.fontSize }]}
        numberOfLines={2}
      >
        {badge.name}
      </Text>
      {!isLocked && userBadge && (
        <Text style={[styles.date, { fontSize: config.fontSize - 2 }]}>
          {new Date(userBadge.awarded_at).toLocaleDateString()}
        </Text>
      )}
      {badge.points_value > 0 && (
        <View style={styles.pointsBadge}>
          <Text style={styles.pointsText}>+{badge.points_value}</Text>
        </View>
      )}
    </TouchableOpacity>
  );
}

const styles = StyleSheet.create({
  container: {
    backgroundColor: '#fff',
    borderRadius: 12,
    alignItems: 'center',
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.05,
    shadowRadius: 2,
    elevation: 2,
  },
  iconContainer: {
    alignItems: 'center',
    justifyContent: 'center',
    marginBottom: 8,
  },
  icon: {
    textAlign: 'center',
  },
  name: {
    fontWeight: '600',
    color: '#1a1a1a',
    textAlign: 'center',
    marginBottom: 2,
  },
  date: {
    color: '#666',
    textAlign: 'center',
  },
  pointsBadge: {
    position: 'absolute',
    top: -4,
    right: -4,
    backgroundColor: '#f59e0b',
    borderRadius: 10,
    paddingHorizontal: 6,
    paddingVertical: 2,
  },
  pointsText: {
    fontSize: 10,
    fontWeight: 'bold',
    color: '#fff',
  },
});
