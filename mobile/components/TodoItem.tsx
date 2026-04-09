import { View, Text, TouchableOpacity, StyleSheet } from 'react-native';
import { Todo, TodoStatus, TodoPriority } from '../types';
import { priorityColors, priorityLabels, statusLabels } from '../stores/todoStore';

interface TodoItemProps {
  todo: Todo | {
    id: string;
    title: string;
    description?: string;
    status: TodoStatus;
    priority: TodoPriority;
    due_date?: string;
  };
  onPress: () => void;
  onToggleComplete: () => void;
}

function formatDueDate(dateString?: string): string | null {
  if (!dateString) return null;
  const date = new Date(dateString);
  const today = new Date();
  const tomorrow = new Date(today);
  tomorrow.setDate(tomorrow.getDate() + 1);

  if (date.toDateString() === today.toDateString()) {
    return 'Today';
  }
  if (date.toDateString() === tomorrow.toDateString()) {
    return 'Tomorrow';
  }
  return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
}

function isOverdue(dateString?: string): boolean {
  if (!dateString) return false;
  const date = new Date(dateString);
  const today = new Date();
  today.setHours(0, 0, 0, 0);
  return date < today;
}

export function TodoItem({ todo, onPress, onToggleComplete }: TodoItemProps) {
  const isCompleted = todo.status === 'completed';
  const dueDate = formatDueDate(todo.due_date);
  const overdue = isOverdue(todo.due_date) && !isCompleted;

  return (
    <TouchableOpacity
      style={[styles.container, isCompleted && styles.completedContainer]}
      onPress={onPress}
      activeOpacity={0.7}
    >
      <TouchableOpacity
        style={[
          styles.checkbox,
          isCompleted && styles.checkboxCompleted,
        ]}
        onPress={onToggleComplete}
        hitSlop={{ top: 10, bottom: 10, left: 10, right: 10 }}
      >
        {isCompleted && <Text style={styles.checkmark}>✓</Text>}
      </TouchableOpacity>

      <View style={styles.content}>
        <Text
          style={[
            styles.title,
            isCompleted && styles.titleCompleted,
          ]}
          numberOfLines={1}
        >
          {todo.title}
        </Text>

        {todo.description && (
          <Text style={styles.description} numberOfLines={1}>
            {todo.description}
          </Text>
        )}

        <View style={styles.meta}>
          <View
            style={[
              styles.priorityBadge,
              { backgroundColor: `${priorityColors[todo.priority]}20` },
            ]}
          >
            <View
              style={[
                styles.priorityDot,
                { backgroundColor: priorityColors[todo.priority] },
              ]}
            />
            <Text
              style={[
                styles.priorityText,
                { color: priorityColors[todo.priority] },
              ]}
            >
              {priorityLabels[todo.priority]}
            </Text>
          </View>

          {dueDate && (
            <Text style={[styles.dueDate, overdue && styles.overdue]}>
              {overdue ? '⚠️ ' : '📅 '}
              {dueDate}
            </Text>
          )}
        </View>
      </View>

      <View style={styles.statusBadge}>
        <Text style={styles.statusText}>{statusLabels[todo.status]}</Text>
      </View>
    </TouchableOpacity>
  );
}

const styles = StyleSheet.create({
  container: {
    flexDirection: 'row',
    alignItems: 'center',
    padding: 16,
    backgroundColor: '#fff',
    borderRadius: 12,
    marginHorizontal: 16,
    marginVertical: 4,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.05,
    shadowRadius: 2,
    elevation: 2,
  },
  completedContainer: {
    opacity: 0.7,
  },
  checkbox: {
    width: 24,
    height: 24,
    borderRadius: 12,
    borderWidth: 2,
    borderColor: '#007AFF',
    marginRight: 12,
    alignItems: 'center',
    justifyContent: 'center',
  },
  checkboxCompleted: {
    backgroundColor: '#007AFF',
  },
  checkmark: {
    color: '#fff',
    fontSize: 14,
    fontWeight: 'bold',
  },
  content: {
    flex: 1,
  },
  title: {
    fontSize: 16,
    fontWeight: '600',
    color: '#1a1a1a',
    marginBottom: 4,
  },
  titleCompleted: {
    textDecorationLine: 'line-through',
    color: '#999',
  },
  description: {
    fontSize: 14,
    color: '#666',
    marginBottom: 8,
  },
  meta: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 12,
  },
  priorityBadge: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 8,
    paddingVertical: 4,
    borderRadius: 12,
    gap: 4,
  },
  priorityDot: {
    width: 6,
    height: 6,
    borderRadius: 3,
  },
  priorityText: {
    fontSize: 12,
    fontWeight: '500',
  },
  dueDate: {
    fontSize: 12,
    color: '#666',
  },
  overdue: {
    color: '#ef4444',
    fontWeight: '500',
  },
  statusBadge: {
    marginLeft: 12,
  },
  statusText: {
    fontSize: 12,
    color: '#999',
  },
});
