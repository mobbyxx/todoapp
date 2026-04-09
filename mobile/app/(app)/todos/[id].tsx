import { useState } from 'react';
import {
  View,
  Text,
  TouchableOpacity,
  StyleSheet,
  ScrollView,
  Alert,
  ActivityIndicator,
} from 'react-native';
import { Stack, useLocalSearchParams, router } from 'expo-router';
import { useTodo, useTodos } from '../../../hooks/useTodos';
import { priorityColors, priorityLabels, statusLabels } from '../../../stores/todoStore';
import type { TodoStatus, TodoPriority } from '../../../types';

export default function TodoDetailScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const [isDeleting, setIsDeleting] = useState(false);
  const [isToggling, setIsToggling] = useState(false);
  const { todo, isLoading } = useTodo(id);
  const { deleteTodo, toggleTodoComplete } = useTodos();

  const handleEdit = () => {
    router.push(`/(app)/todos/edit/${id}`);
  };

  const handleDelete = () => {
    Alert.alert(
      'Delete Todo',
      'Are you sure you want to delete this todo? This action cannot be undone.',
      [
        { text: 'Cancel', style: 'cancel' },
        {
          text: 'Delete',
          style: 'destructive',
          onPress: async () => {
            setIsDeleting(true);
            try {
              await deleteTodo(id);
              router.back();
            } catch {
              Alert.alert('Error', 'Failed to delete todo');
            } finally {
              setIsDeleting(false);
            }
          },
        },
      ]
    );
  };

  const handleToggleComplete = async () => {
    if (isToggling) return;
    setIsToggling(true);
    try {
      await toggleTodoComplete(id);
    } catch {
      Alert.alert('Error', 'Failed to update todo status');
    } finally {
      setIsToggling(false);
    }
  };

  if (isLoading) {
    return (
      <View style={styles.loadingContainer}>
        <ActivityIndicator size="large" color="#007AFF" />
      </View>
    );
  }

  if (!todo) {
    return (
      <View style={styles.errorContainer}>
        <Text style={styles.errorText}>Todo not found</Text>
        <TouchableOpacity
          style={styles.backButton}
          onPress={() => router.back()}
        >
          <Text style={styles.backButtonText}>Go Back</Text>
        </TouchableOpacity>
      </View>
    );
  }

  const todoData = todo.toJSON();
  const isCompleted = todoData.status === 'completed';
  const formattedDueDate = todoData.due_date
    ? new Date(todoData.due_date).toLocaleDateString('en-US', {
        weekday: 'long',
        year: 'numeric',
        month: 'long',
        day: 'numeric',
      })
    : null;

  const formattedCreatedAt = new Date(todoData.created_at).toLocaleDateString(
    'en-US',
    {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    }
  );

  return (
    <View style={styles.container}>
      <Stack.Screen
        options={{
          title: 'Todo Details',
          headerRight: () => (
            <View style={styles.headerButtons}>
              <TouchableOpacity
                style={styles.headerButton}
                onPress={handleEdit}
                disabled={isToggling || isDeleting}
              >
                <Text style={styles.headerButtonText}>Edit</Text>
              </TouchableOpacity>
            </View>
          ),
        }}
      />

      <ScrollView style={styles.content} showsVerticalScrollIndicator={false}>
        <View style={styles.header}>
          <View
            style={[
              styles.priorityBadge,
              { backgroundColor: `${priorityColors[todoData.priority as TodoPriority]}20` },
            ]}
          >
            <View
              style={[
                styles.priorityDot,
                { backgroundColor: priorityColors[todoData.priority as TodoPriority] },
              ]}
            />
            <Text
              style={[
                styles.priorityText,
                { color: priorityColors[todoData.priority as TodoPriority] },
              ]}
            >
              {priorityLabels[todoData.priority as TodoPriority]} Priority
            </Text>
          </View>

          <View
            style={[
              styles.statusBadge,
              isCompleted && styles.statusBadgeCompleted,
            ]}
          >
            <Text
              style={[
                styles.statusText,
                isCompleted && styles.statusTextCompleted,
              ]}
            >
              {statusLabels[todoData.status as TodoStatus]}
            </Text>
          </View>
        </View>

        <View style={styles.section}>
          <Text
            style={[styles.title, isCompleted && styles.titleCompleted]}
          >
            {todoData.title}
          </Text>

          {todoData.description && (
            <Text style={styles.description}>{todoData.description}</Text>
          )}
        </View>

        <View style={styles.metaSection}>
          {formattedDueDate && (
            <View style={styles.metaItem}>
              <Text style={styles.metaLabel}>Due Date</Text>
              <Text style={styles.metaValue}>📅 {formattedDueDate}</Text>
            </View>
          )}

          <View style={styles.metaItem}>
            <Text style={styles.metaLabel}>Created</Text>
            <Text style={styles.metaValue}>{formattedCreatedAt}</Text>
          </View>

          {todoData.assigned_to && (
            <View style={styles.metaItem}>
              <Text style={styles.metaLabel}>Assigned To</Text>
              <Text style={styles.metaValue}>👤 Connected Partner</Text>
            </View>
          )}

          {todoData.completed_at && (
            <View style={styles.metaItem}>
              <Text style={styles.metaLabel}>Completed</Text>
              <Text style={styles.metaValue}>
                ✅{' '}
                {new Date(todoData.completed_at).toLocaleDateString('en-US', {
                  year: 'numeric',
                  month: 'short',
                  day: 'numeric',
                  hour: '2-digit',
                  minute: '2-digit',
                })}
              </Text>
            </View>
          )}
        </View>

        <View style={styles.actionsSection}>
          <TouchableOpacity
            style={[
              styles.actionButton,
              isCompleted ? styles.actionButtonPending : styles.actionButtonComplete,
              (isToggling || isDeleting) && styles.actionButtonDisabled,
            ]}
            onPress={handleToggleComplete}
            disabled={isToggling || isDeleting}
          >
            {isToggling ? (
              <ActivityIndicator color="#fff" />
            ) : (
              <Text style={styles.actionButtonText}>
                {isCompleted ? 'Mark as Pending' : 'Mark as Complete'}
              </Text>
            )}
          </TouchableOpacity>

          <TouchableOpacity
            style={[
              styles.actionButton,
              styles.actionButtonDelete,
              isDeleting && styles.actionButtonDisabled,
            ]}
            onPress={handleDelete}
            disabled={isToggling || isDeleting}
          >
            {isDeleting ? (
              <ActivityIndicator color="#dc2626" />
            ) : (
              <Text style={styles.actionButtonTextDelete}>Delete Todo</Text>
            )}
          </TouchableOpacity>
        </View>
      </ScrollView>
    </View>
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
  },
  errorContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 20,
  },
  errorText: {
    fontSize: 18,
    color: '#666',
    marginBottom: 20,
  },
  backButton: {
    backgroundColor: '#007AFF',
    paddingHorizontal: 24,
    paddingVertical: 12,
    borderRadius: 8,
  },
  backButtonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '600',
  },
  headerButtons: {
    flexDirection: 'row',
    gap: 8,
  },
  headerButton: {
    paddingHorizontal: 12,
    paddingVertical: 6,
  },
  headerButtonText: {
    color: '#007AFF',
    fontSize: 16,
    fontWeight: '600',
  },
  content: {
    flex: 1,
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: 20,
    backgroundColor: '#fff',
    borderBottomWidth: 1,
    borderBottomColor: '#f0f0f0',
  },
  priorityBadge: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 12,
    paddingVertical: 6,
    borderRadius: 16,
    gap: 6,
  },
  priorityDot: {
    width: 8,
    height: 8,
    borderRadius: 4,
  },
  priorityText: {
    fontSize: 14,
    fontWeight: '600',
  },
  statusBadge: {
    paddingHorizontal: 12,
    paddingVertical: 6,
    borderRadius: 16,
    backgroundColor: '#f5f5f5',
  },
  statusBadgeCompleted: {
    backgroundColor: '#f0fdf4',
  },
  statusText: {
    fontSize: 14,
    fontWeight: '600',
    color: '#666',
  },
  statusTextCompleted: {
    color: '#16a34a',
  },
  section: {
    backgroundColor: '#fff',
    padding: 20,
    marginBottom: 12,
  },
  title: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#1a1a1a',
    marginBottom: 12,
  },
  titleCompleted: {
    textDecorationLine: 'line-through',
    color: '#999',
  },
  description: {
    fontSize: 16,
    color: '#666',
    lineHeight: 24,
  },
  metaSection: {
    backgroundColor: '#fff',
    padding: 20,
    marginBottom: 12,
  },
  metaItem: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    paddingVertical: 12,
    borderBottomWidth: 1,
    borderBottomColor: '#f0f0f0',
  },
  metaLabel: {
    fontSize: 14,
    color: '#666',
  },
  metaValue: {
    fontSize: 14,
    color: '#1a1a1a',
    fontWeight: '500',
  },
  actionsSection: {
    backgroundColor: '#fff',
    padding: 20,
    gap: 12,
    marginBottom: 40,
  },
  actionButton: {
    paddingVertical: 16,
    borderRadius: 12,
    alignItems: 'center',
  },
  actionButtonComplete: {
    backgroundColor: '#22c55e',
  },
  actionButtonPending: {
    backgroundColor: '#eab308',
  },
  actionButtonDelete: {
    backgroundColor: '#fef2f2',
    borderWidth: 1,
    borderColor: '#fecaca',
  },
  actionButtonDisabled: {
    opacity: 0.6,
  },
  actionButtonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '600',
  },
  actionButtonTextDelete: {
    color: '#dc2626',
    fontSize: 16,
    fontWeight: '600',
  },
});
