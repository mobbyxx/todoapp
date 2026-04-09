import { useState } from 'react';
import {
  View,
  StyleSheet,
  KeyboardAvoidingView,
  Platform,
  Alert,
} from 'react-native';
import { Stack, router } from 'expo-router';
import { useTodos } from '../../../hooks/useTodos';
import { TodoCreateInput, TodoUpdateInput } from '../../../types';
import { TodoForm } from '../../../components/TodoForm';

export default function CreateTodoScreen() {
  const [connections] = useState<{ id: string; display_name: string }[]>([]);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const { createTodo } = useTodos();

  const handleSubmit = async (data: TodoCreateInput | TodoUpdateInput) => {
    if (isSubmitting) return;
    
    setIsSubmitting(true);
    try {
      const createData = data as TodoCreateInput;
      await createTodo({
        title: createData.title,
        description: createData.description,
        priority: createData.priority,
        assignedTo: createData.assigned_to,
        dueDate: createData.due_date,
        tags: createData.tags,
      });
      router.back();
    } catch (error) {
      Alert.alert('Error', 'Failed to create todo');
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleCancel = () => {
    router.back();
  };

  return (
    <KeyboardAvoidingView
      behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
      style={styles.container}
    >
      <Stack.Screen
        options={{
          title: 'New Todo',
          headerLeft: () => null,
        }}
      />

      <View style={styles.content}>
        <TodoForm
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
});
