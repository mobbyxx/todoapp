import { useState, useCallback } from 'react';
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  RefreshControl,
  ActivityIndicator,
} from 'react-native';
import { Stack } from 'expo-router';
import { useQuery } from '@tanstack/react-query';
import {
  getUserProfile,
  getUserBadges,
  getLevelProgress,
  getCurrentStreak,
} from '../../services/gamification';
import { ProgressBar } from '../../components/gamification/ProgressBar';
import { BadgeCard } from '../../components/gamification/BadgeCard';

export default function StatsScreen() {
  const [isRefreshing, setIsRefreshing] = useState(false);

  const { data: profile, isLoading: isLoadingProfile, refetch: refetchProfile } = useQuery({
    queryKey: ['userProfile'],
    queryFn: getUserProfile,
  });

  const { data: badges, isLoading: isLoadingBadges, refetch: refetchBadges } = useQuery({
    queryKey: ['userBadges'],
    queryFn: getUserBadges,
  });

  const { data: levelProgress, isLoading: isLoadingProgress, refetch: refetchProgress } = useQuery({
    queryKey: ['levelProgress'],
    queryFn: getLevelProgress,
  });

  const { data: streak, isLoading: isLoadingStreak, refetch: refetchStreak } = useQuery({
    queryKey: ['streak'],
    queryFn: getCurrentStreak,
  });

  const isLoading = isLoadingProfile || isLoadingBadges || isLoadingProgress || isLoadingStreak;

  const handleRefresh = useCallback(async () => {
    setIsRefreshing(true);
    try {
      await Promise.all([
        refetchProfile(),
        refetchBadges(),
        refetchProgress(),
        refetchStreak(),
      ]);
    } finally {
      setIsRefreshing(false);
    }
  }, [refetchProfile, refetchBadges, refetchProgress, refetchStreak]);

  if (isLoading && !isRefreshing) {
    return (
      <View style={styles.loadingContainer}>
        <ActivityIndicator size="large" color="#3b82f6" />
        <Text style={styles.loadingText}>Loading stats...</Text>
      </View>
    );
  }

  const level = profile?.current_level;
  const userBadges = badges || [];
  const progress = levelProgress || { currentPoints: 0, nextLevelPoints: 100, progressPercentage: 0 };
  const streakData = streak || { currentStreak: 0, longestStreak: 0, lastActiveDate: '' };

  return (
    <ScrollView
      style={styles.container}
      refreshControl={
        <RefreshControl
          refreshing={isRefreshing}
          onRefresh={handleRefresh}
          tintColor="#3b82f6"
        />
      }
    >
      <Stack.Screen options={{ title: 'My Stats' }} />

      <View style={styles.statsGrid}>
        <View style={styles.statCard}>
          <Text style={styles.statValue}>{profile?.total_points || 0}</Text>
          <Text style={styles.statLabel}>Total Points</Text>
        </View>
        <View style={styles.statCard}>
          <Text style={styles.statValue}>{streakData.currentStreak}</Text>
          <Text style={styles.statLabel}>Day Streak</Text>
        </View>
      </View>

      <View style={styles.levelCard}>
        <View style={styles.levelHeader}>
          <View style={styles.levelBadge}>
            <Text style={styles.levelIcon}>🏆</Text>
            <Text style={styles.levelNumber}>{level?.level_number || 1}</Text>
          </View>
          <View style={styles.levelInfo}>
            <Text style={styles.levelName}>{level?.name || 'Beginner'}</Text>
            <Text style={styles.levelSubtitle}>
              {progress.currentPoints} / {progress.nextLevelPoints} XP
            </Text>
          </View>
        </View>

        <View style={styles.progressContainer}>
          <ProgressBar
            progress={progress.progressPercentage}
            height={10}
            animated={true}
            showGlow={progress.progressPercentage >= 90}
          />
          <Text style={styles.progressText}>{Math.round(progress.progressPercentage)}% to next level</Text>
        </View>
      </View>

      <View style={styles.streakCard}>
        <Text style={styles.sectionTitle}>🔥 Streak</Text>
        <View style={styles.streakGrid}>
          <View style={styles.streakItem}>
            <Text style={styles.streakValue}>{streakData.currentStreak}</Text>
            <Text style={styles.streakLabel}>Current</Text>
          </View>
          <View style={styles.streakItem}>
            <Text style={styles.streakValue}>{streakData.longestStreak}</Text>
            <Text style={styles.streakLabel}>Longest</Text>
          </View>
        </View>
      </View>

      <View style={styles.statsSummaryCard}>
        <Text style={styles.sectionTitle}>📊 Activity</Text>
        <View style={styles.activityRow}>
          <Text style={styles.activityLabel}>Total Todos</Text>
          <Text style={styles.activityValue}>{profile?.total_todos || 0}</Text>
        </View>
        <View style={styles.activityRow}>
          <Text style={styles.activityLabel}>Completed</Text>
          <Text style={styles.activityValue}>{profile?.completed_todos || 0}</Text>
        </View>
        <View style={styles.activityRow}>
          <Text style={styles.activityLabel}>Completion Rate</Text>
          <Text style={styles.activityValue}>
            {profile?.total_todos
              ? Math.round(((profile.completed_todos || 0) / profile.total_todos) * 100)
              : 0}%
          </Text>
        </View>
      </View>

      <View style={styles.badgesSection}>
        <Text style={styles.sectionTitle}>🏅 Badges ({userBadges.length})</Text>
        {userBadges.length === 0 ? (
          <View style={styles.emptyBadges}>
            <Text style={styles.emptyEmoji}>🏆</Text>
            <Text style={styles.emptyTitle}>No Badges Yet</Text>
            <Text style={styles.emptyText}>Complete tasks to earn badges!</Text>
          </View>
        ) : (
          <View style={styles.badgesGrid}>
            {userBadges.map((userBadge) => (
              <BadgeCard
                key={userBadge.id}
                badge={userBadge.badge}
                userBadge={userBadge}
                size="medium"
              />
            ))}
          </View>
        )}
      </View>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f5f5f5',
  },
  loadingContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: '#f5f5f5',
  },
  loadingText: {
    marginTop: 12,
    fontSize: 16,
    color: '#666',
  },
  statsGrid: {
    flexDirection: 'row',
    padding: 16,
    gap: 12,
  },
  statCard: {
    flex: 1,
    backgroundColor: '#fff',
    borderRadius: 16,
    padding: 20,
    alignItems: 'center',
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.05,
    shadowRadius: 4,
    elevation: 2,
  },
  statValue: {
    fontSize: 32,
    fontWeight: 'bold',
    color: '#3b82f6',
    marginBottom: 4,
  },
  statLabel: {
    fontSize: 14,
    color: '#666',
    fontWeight: '500',
  },
  levelCard: {
    backgroundColor: '#fff',
    borderRadius: 16,
    padding: 20,
    marginHorizontal: 16,
    marginBottom: 12,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.05,
    shadowRadius: 4,
    elevation: 2,
  },
  levelHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 16,
  },
  levelBadge: {
    width: 56,
    height: 56,
    borderRadius: 28,
    backgroundColor: '#fbbf24',
    alignItems: 'center',
    justifyContent: 'center',
    marginRight: 16,
    borderWidth: 3,
    borderColor: '#f59e0b',
  },
  levelIcon: {
    fontSize: 24,
    marginBottom: -4,
  },
  levelNumber: {
    fontSize: 16,
    fontWeight: 'bold',
    color: '#fff',
  },
  levelInfo: {
    flex: 1,
  },
  levelName: {
    fontSize: 20,
    fontWeight: 'bold',
    color: '#1a1a1a',
    marginBottom: 2,
  },
  levelSubtitle: {
    fontSize: 14,
    color: '#666',
  },
  progressContainer: {
    gap: 8,
  },
  progressText: {
    fontSize: 12,
    color: '#666',
    textAlign: 'center',
  },
  streakCard: {
    backgroundColor: '#fff',
    borderRadius: 16,
    padding: 20,
    marginHorizontal: 16,
    marginBottom: 12,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.05,
    shadowRadius: 4,
    elevation: 2,
  },
  sectionTitle: {
    fontSize: 18,
    fontWeight: 'bold',
    color: '#1a1a1a',
    marginBottom: 16,
  },
  streakGrid: {
    flexDirection: 'row',
    gap: 16,
  },
  streakItem: {
    flex: 1,
    alignItems: 'center',
    padding: 16,
    backgroundColor: '#fef3c7',
    borderRadius: 12,
  },
  streakValue: {
    fontSize: 28,
    fontWeight: 'bold',
    color: '#f59e0b',
    marginBottom: 4,
  },
  streakLabel: {
    fontSize: 14,
    color: '#666',
    fontWeight: '500',
  },
  statsSummaryCard: {
    backgroundColor: '#fff',
    borderRadius: 16,
    padding: 20,
    marginHorizontal: 16,
    marginBottom: 12,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.05,
    shadowRadius: 4,
    elevation: 2,
  },
  activityRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    paddingVertical: 12,
    borderBottomWidth: 1,
    borderBottomColor: '#f0f0f0',
  },
  activityLabel: {
    fontSize: 14,
    color: '#666',
  },
  activityValue: {
    fontSize: 14,
    fontWeight: '600',
    color: '#1a1a1a',
  },
  badgesSection: {
    padding: 16,
    paddingBottom: 32,
  },
  emptyBadges: {
    backgroundColor: '#fff',
    borderRadius: 16,
    padding: 40,
    alignItems: 'center',
  },
  emptyEmoji: {
    fontSize: 48,
    marginBottom: 16,
  },
  emptyTitle: {
    fontSize: 18,
    fontWeight: '600',
    color: '#1a1a1a',
    marginBottom: 8,
  },
  emptyText: {
    fontSize: 14,
    color: '#666',
    textAlign: 'center',
  },
  badgesGrid: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 12,
    justifyContent: 'flex-start',
  },
});
