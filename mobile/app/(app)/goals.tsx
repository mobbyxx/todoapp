import { useState, useCallback } from 'react';
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  RefreshControl,
  ActivityIndicator,
  Alert,
} from 'react-native';
import { Stack } from 'expo-router';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { getActiveGoals, getGoals, joinGoal, Goal } from '../../services/gamification';
import { GoalCard } from '../../components/gamification/GoalCard';
import { useAuthStore } from '../../stores/authStore';

export default function GoalsScreen() {
  const queryClient = useQueryClient();
  const { user } = useAuthStore();
  const [isRefreshing, setIsRefreshing] = useState(false);

  const { data: goals, isLoading, refetch } = useQuery({
    queryKey: ['goals'],
    queryFn: getGoals,
  });

  const { data: activeGoals, refetch: refetchActive } = useQuery({
    queryKey: ['activeGoals'],
    queryFn: getActiveGoals,
  });

  const joinMutation = useMutation({
    mutationFn: joinGoal,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['goals'] });
      queryClient.invalidateQueries({ queryKey: ['activeGoals'] });
      Alert.alert('Success', 'You have joined the goal!');
    },
    onError: (error: Error) => {
      Alert.alert('Error', error.message || 'Failed to join goal');
    },
  });

  const handleRefresh = useCallback(async () => {
    setIsRefreshing(true);
    try {
      await Promise.all([refetch(), refetchActive()]);
    } finally {
      setIsRefreshing(false);
    }
  }, [refetch, refetchActive]);

  const handleJoinGoal = useCallback((goal: Goal) => {
    Alert.alert(
      'Join Goal',
      `Join "${goal.title}"?`,
      [
        { text: 'Cancel', style: 'cancel' },
        { text: 'Join', onPress: () => joinMutation.mutate(goal.id) },
      ]
    );
  }, [joinMutation]);

  const allGoals = goals || [];
  const activeGoalsList = activeGoals || [];

  const userGoalIds = new Set(
    allGoals
      .filter((g) => user && g.participants.includes(user.id))
      .map((g) => g.id)
  );

  const myGoals = allGoals.filter((g) => userGoalIds.has(g.id));
  const availableGoals = allGoals.filter((g) => !userGoalIds.has(g.id) && g.is_active);

  if (isLoading && !isRefreshing) {
    return (
      <View style={styles.loadingContainer}>
        <ActivityIndicator size="large" color="#3b82f6" />
        <Text style={styles.loadingText}>Loading goals...</Text>
      </View>
    );
  }

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
      <Stack.Screen options={{ title: 'Shared Goals' }} />

      <View style={styles.statsCard}>
        <View style={styles.statItem}>
          <Text style={styles.statValue}>{myGoals.length}</Text>
          <Text style={styles.statLabel}>My Goals</Text>
        </View>
        <View style={styles.statDivider} />
        <View style={styles.statItem}>
          <Text style={styles.statValue}>{activeGoalsList.length}</Text>
          <Text style={styles.statLabel}>Active</Text>
        </View>
        <View style={styles.statDivider} />
        <View style={styles.statItem}>
          <Text style={styles.statValue}>
            {myGoals.filter((g) => g.current_value >= g.target_value).length}
          </Text>
          <Text style={styles.statLabel}>Completed</Text>
        </View>
      </View>

      {myGoals.length > 0 && (
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>🎯 My Goals</Text>
          {myGoals.map((goal) => (
            <GoalCard
              key={goal.id}
              goal={goal}
              isParticipant={true}
              showProgress={true}
            />
          ))}
        </View>
      )}

      {availableGoals.length > 0 && (
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>🌟 Available Goals</Text>
          <Text style={styles.sectionSubtitle}>Join these goals to collaborate with others</Text>
          {availableGoals.map((goal) => (
            <GoalCard
              key={goal.id}
              goal={goal}
              onJoin={handleJoinGoal}
              isParticipant={false}
              showProgress={true}
            />
          ))}
        </View>
      )}

      {myGoals.length === 0 && availableGoals.length === 0 && (
        <View style={styles.emptyContainer}>
          <Text style={styles.emptyEmoji}>🎯</Text>
          <Text style={styles.emptyTitle}>No Goals Yet</Text>
          <Text style={styles.emptyText}>
            Shared goals will appear here when they are created.
          </Text>
        </View>
      )}
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
  statsCard: {
    flexDirection: 'row',
    backgroundColor: '#fff',
    marginHorizontal: 16,
    marginTop: 16,
    marginBottom: 8,
    borderRadius: 16,
    padding: 20,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.05,
    shadowRadius: 4,
    elevation: 2,
  },
  statItem: {
    flex: 1,
    alignItems: 'center',
  },
  statDivider: {
    width: 1,
    backgroundColor: '#e5e7eb',
  },
  statValue: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#3b82f6',
    marginBottom: 4,
  },
  statLabel: {
    fontSize: 13,
    color: '#666',
    fontWeight: '500',
  },
  section: {
    marginTop: 8,
    paddingBottom: 8,
  },
  sectionTitle: {
    fontSize: 18,
    fontWeight: 'bold',
    color: '#1a1a1a',
    marginHorizontal: 16,
    marginBottom: 12,
    marginTop: 8,
  },
  sectionSubtitle: {
    fontSize: 13,
    color: '#666',
    marginHorizontal: 16,
    marginBottom: 12,
  },
  emptyContainer: {
    paddingVertical: 80,
    alignItems: 'center',
    paddingHorizontal: 40,
  },
  emptyEmoji: {
    fontSize: 56,
    marginBottom: 16,
  },
  emptyTitle: {
    fontSize: 20,
    fontWeight: '600',
    color: '#1a1a1a',
    marginBottom: 8,
  },
  emptyText: {
    fontSize: 14,
    color: '#666',
    textAlign: 'center',
  },
});
