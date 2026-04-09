import { View, Text, StyleSheet, TouchableOpacity } from 'react-native';
import { Reward, UserReward, RewardType } from '../../types';

interface RewardCardProps {
  reward: Reward;
  userReward?: UserReward;
  onRedeem?: (reward: Reward) => void;
  showRedeemButton?: boolean;
}

const rewardTypeIcons: Record<RewardType, string> = {
  badge: '🏅',
  points: '💎',
  feature: '✨',
};

const rewardTypeColors: Record<RewardType, string> = {
  badge: '#8b5cf6',
  points: '#f59e0b',
  feature: '#10b981',
};

export function RewardCard({
  reward,
  userReward,
  onRedeem,
  showRedeemButton = true,
}: RewardCardProps) {
  const isRedeemed = !!userReward;
  const icon = rewardTypeIcons[reward.type];
  const color = rewardTypeColors[reward.type];

  return (
    <View style={[styles.container, isRedeemed && styles.redeemedContainer]}>
      <View style={[styles.iconContainer, { backgroundColor: `${color}20` }]}>
        <Text style={styles.icon}>{icon}</Text>
      </View>

      <View style={styles.content}>
        <Text style={styles.name}>{reward.name}</Text>
        <Text style={styles.description} numberOfLines={2}>
          {reward.description}
        </Text>

        <View style={styles.meta}>
          <View style={[styles.typeBadge, { backgroundColor: `${color}20` }]}>
            <Text style={[styles.typeText, { color }]}>
              {reward.type.charAt(0).toUpperCase() + reward.type.slice(1)}
            </Text>
          </View>
          <Text style={styles.value}>💎 {reward.value}</Text>
        </View>

        {isRedeemed && userReward && (
          <Text style={styles.redeemedDate}>
            Redeemed {new Date(userReward.claimed_at).toLocaleDateString()}
            {userReward.expires_at &&
              ` • Expires ${new Date(userReward.expires_at).toLocaleDateString()}`}
          </Text>
        )}
      </View>

      {showRedeemButton && !isRedeemed && (
        <TouchableOpacity
          style={[styles.redeemButton, { backgroundColor: color }]}
          onPress={() => onRedeem?.(reward)}
          activeOpacity={0.8}
        >
          <Text style={styles.redeemButtonText}>Redeem</Text>
        </TouchableOpacity>
      )}

      {isRedeemed && (
        <View style={styles.redeemedBadge}>
          <Text style={styles.redeemedBadgeText}>✓</Text>
        </View>
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: '#fff',
    borderRadius: 12,
    padding: 16,
    marginHorizontal: 16,
    marginVertical: 6,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.05,
    shadowRadius: 2,
    elevation: 2,
  },
  redeemedContainer: {
    opacity: 0.8,
    backgroundColor: '#f9fafb',
  },
  iconContainer: {
    width: 48,
    height: 48,
    borderRadius: 24,
    alignItems: 'center',
    justifyContent: 'center',
    marginRight: 12,
  },
  icon: {
    fontSize: 24,
  },
  content: {
    flex: 1,
  },
  name: {
    fontSize: 16,
    fontWeight: '600',
    color: '#1a1a1a',
    marginBottom: 4,
  },
  description: {
    fontSize: 13,
    color: '#666',
    marginBottom: 8,
    lineHeight: 18,
  },
  meta: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
  },
  typeBadge: {
    paddingHorizontal: 8,
    paddingVertical: 4,
    borderRadius: 12,
  },
  typeText: {
    fontSize: 11,
    fontWeight: '600',
  },
  value: {
    fontSize: 13,
    fontWeight: '600',
    color: '#666',
  },
  redeemedDate: {
    fontSize: 12,
    color: '#999',
    marginTop: 4,
  },
  redeemButton: {
    paddingHorizontal: 16,
    paddingVertical: 8,
    borderRadius: 8,
    marginLeft: 12,
  },
  redeemButtonText: {
    color: '#fff',
    fontSize: 13,
    fontWeight: '600',
  },
  redeemedBadge: {
    width: 24,
    height: 24,
    borderRadius: 12,
    backgroundColor: '#22c55e',
    alignItems: 'center',
    justifyContent: 'center',
    marginLeft: 12,
  },
  redeemedBadgeText: {
    color: '#fff',
    fontSize: 14,
    fontWeight: 'bold',
  },
});
