import { View, Text, StyleSheet, TouchableOpacity } from 'react-native';
import { Goal } from '../../services/gamification';
import { ProgressBar } from './ProgressBar';

interface GoalCardProps {
  goal: Goal;
  onJoin?: (goal: Goal) => void;
  isParticipant?: boolean;
  showProgress?: boolean;
}

export function GoalCard({
  goal,
  onJoin,
  isParticipant = false,
  showProgress = true,
}: GoalCardProps) {
  const progress = Math.min(
    100,
    Math.round((goal.current_value / goal.target_value) * 100)
  );
  const isCompleted = goal.current_value >= goal.target_value;

  const formatDeadline = (dateString?: string): string | null => {
    if (!dateString) return null;
    const date = new Date(dateString);
    const today = new Date();
    const daysLeft = Math.ceil((date.getTime() - today.getTime()) / (1000 * 60 * 60 * 24));

    if (daysLeft < 0) return 'Overdue';
    if (daysLeft === 0) return 'Due today';
    if (daysLeft === 1) return '1 day left';
    return `${daysLeft} days left`;
  };

  return (
    <View style={[styles.container, isCompleted && styles.completedContainer]}>
      <View style={styles.header}>
        <View style={styles.titleContainer}>
          <Text style={styles.title}>{goal.title}</Text>
          {isCompleted && (
            <View style={styles.completedBadge}>
              <Text style={styles.completedBadgeText}>✓</Text>
            </View>
          )}
        </View>
        {goal.description && (
          <Text style={styles.description} numberOfLines={2}>
            {goal.description}
          </Text>
        )}
      </View>

      {showProgress && (
        <View style={styles.progressSection}>
          <View style={styles.progressHeader}>
            <Text style={styles.progressText}>
              {goal.current_value} / {goal.target_value} {goal.unit}
            </Text>
            <Text
              style={[
                styles.progressPercentage,
                isCompleted && styles.completedPercentage,
              ]}
            >
              {progress}%
            </Text>
          </View>
          <ProgressBar
            progress={progress}
            height={6}
            animated={true}
            showGlow={progress >= 90}
          />
        </View>
      )}

      <View style={styles.footer}>
        <View style={styles.meta}>
          <Text style={styles.participants}>
            👥 {goal.participants.length} participant
            {goal.participants.length !== 1 ? 's' : ''}
          </Text>
          {goal.deadline && (
            <Text
              style={[
                styles.deadline,
                formatDeadline(goal.deadline)?.includes('Overdue') &&
                  styles.deadlineOverdue,
              ]}
            >
              📅 {formatDeadline(goal.deadline)}
            </Text>
          )}
        </View>

        {!isParticipant && onJoin && !isCompleted && (
          <TouchableOpacity
            style={styles.joinButton}
            onPress={() => onJoin(goal)}
            activeOpacity={0.8}
          >
            <Text style={styles.joinButtonText}>Join</Text>
          </TouchableOpacity>
        )}

        {isParticipant && (
          <View style={styles.participantBadge}>
            <Text style={styles.participantBadgeText}>Joined</Text>
          </View>
        )}
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
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
  completedContainer: {
    borderColor: '#22c55e',
    borderWidth: 1,
  },
  header: {
    marginBottom: 12,
  },
  titleContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
  },
  title: {
    fontSize: 16,
    fontWeight: '600',
    color: '#1a1a1a',
    flex: 1,
  },
  completedBadge: {
    width: 20,
    height: 20,
    borderRadius: 10,
    backgroundColor: '#22c55e',
    alignItems: 'center',
    justifyContent: 'center',
  },
  completedBadgeText: {
    color: '#fff',
    fontSize: 12,
    fontWeight: 'bold',
  },
  description: {
    fontSize: 13,
    color: '#666',
    marginTop: 4,
    lineHeight: 18,
  },
  progressSection: {
    marginBottom: 12,
  },
  progressHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 8,
  },
  progressText: {
    fontSize: 13,
    color: '#666',
  },
  progressPercentage: {
    fontSize: 13,
    fontWeight: '600',
    color: '#3b82f6',
  },
  completedPercentage: {
    color: '#22c55e',
  },
  footer: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  meta: {
    flex: 1,
  },
  participants: {
    fontSize: 12,
    color: '#666',
    marginBottom: 2,
  },
  deadline: {
    fontSize: 12,
    color: '#666',
  },
  deadlineOverdue: {
    color: '#ef4444',
    fontWeight: '500',
  },
  joinButton: {
    backgroundColor: '#3b82f6',
    paddingHorizontal: 16,
    paddingVertical: 8,
    borderRadius: 8,
  },
  joinButtonText: {
    color: '#fff',
    fontSize: 13,
    fontWeight: '600',
  },
  participantBadge: {
    backgroundColor: '#dbeafe',
    paddingHorizontal: 12,
    paddingVertical: 6,
    borderRadius: 8,
  },
  participantBadgeText: {
    color: '#3b82f6',
    fontSize: 12,
    fontWeight: '600',
  },
});
