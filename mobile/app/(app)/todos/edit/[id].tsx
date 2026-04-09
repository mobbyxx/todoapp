import { useState } from 'react';
import {
  View,
  StyleSheet,
  KeyboardAvoidingView,
  Platform,
  Alert,
  ActivityIndicator,
  Text,
} from 'react-native';
import { Stack, useLocalSearchParams, router } from 'expo-router';
import { useTodo, useTodos } from '../../../../hooks/useTodos';
import { TodoCreateInput, TodoUpdateInput } from '../../../../types';
import { TodoForm } from '../../../../components/TodoForm';

export default function EditTodoScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const [connections] = useState<{ id: string; display_name: string }[]>([]);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const { todo, isLoading: isLoadingTodo } = useTodo(id);
  const { updateTodo } = useTodos();

  const handleSubmit = async (data: TodoCreateInput | TodoUpdateInput) => {
    if (isSubmitting) return;
    
    setIsSubmitting(true);
    try {
      const updateData = data as TodoUpdateInput;
      await updateTodo(id, {
        title: updateData.title,
        description: updateData.description,
        status: updateData.status,
        priority: updateData.priority,
        assignedTo: updateData.assigned_to,
        dueDate: updateData.due_date,
        tags: updateData.tags,
      });
      router.back();
    } catch (error) {
      Alert.alert('Error', 'Failed to update todo');
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleCancel = () => {
    router.back();
  };

  if (isLoadingTodo) {
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
      </View>
    );
  }

  const todoData = todo.toJSON();
  const todoFormData = {
    ...todoData,
    description: todoData.description || undefined,
    assigned_to: todoData.assigned_to || undefined,
    due_date: todoData.due_date || undefined,
    tags: todoData.tags || undefined,
  };

  return (
    <KeyboardAvoidingView
      behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
      style={styles.container}
    >
      <Stack.Screen
        options={{
          title: 'Edit Todo',
          headerLeft: () => null,
        }}
      />

      <View style={styles.content}>
        <TodoForm
          todo={todoFormData}
          connections={connections}
          onSubmit={handleSubmit}
          onCancel={handleCancel}
          isLoading={isSubmitting}
        />
      </View>
    </KeyboardAvoidingView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f5f5f5',
  },
  content: {
    flex: 1,
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
  },
});
