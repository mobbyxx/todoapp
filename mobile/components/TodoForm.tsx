import { useState } from 'react';
import {
  View,
  Text,
  TextInput,
  TouchableOpacity,
  StyleSheet,
  ScrollView,
} from 'react-native';
import { Todo, TodoCreateInput, TodoUpdateInput, TodoPriority, TodoStatus } from '../types';
import { priorityColors, priorityLabels } from '../stores/todoStore';

interface TodoFormProps {
  todo?: Todo;
  connections: { id: string; display_name: string }[];
  onSubmit: (data: TodoCreateInput | TodoUpdateInput) => void;
  onCancel: () => void;
  isLoading?: boolean;
}

interface FormErrors {
  title?: string;
  description?: string;
  due_date?: string;
}

const priorities: TodoPriority[] = ['low', 'medium', 'high', 'urgent'];
const statuses: TodoStatus[] = ['pending', 'in_progress', 'completed'];

function validateTitle(title: string): string | undefined {
  if (!title.trim()) return 'Title is required';
  if (title.length > 200) return 'Title must be less than 200 characters';
  return undefined;
}

function validateDescription(description: string): string | undefined {
  if (description.length > 2000) return 'Description must be less than 2000 characters';
  return undefined;
}

export function TodoForm({
  todo,
  connections,
  onSubmit,
  onCancel,
  isLoading = false,
}: TodoFormProps) {
  const [title, setTitle] = useState(todo?.title || '');
  const [description, setDescription] = useState(todo?.description || '');
  const [priority, setPriority] = useState<TodoPriority>(todo?.priority || 'medium');
  const [status, setStatus] = useState<TodoStatus>(todo?.status || 'pending');
  const [assignedTo, setAssignedTo] = useState(todo?.assigned_to || '');
  const [dueDate, setDueDate] = useState(todo?.due_date || '');
  const [errors, setErrors] = useState<FormErrors>({});

  const isEditing = !!todo;

  const validateForm = (): boolean => {
    const newErrors: FormErrors = {};
    const titleError = validateTitle(title);
    const descriptionError = validateDescription(description);

    if (titleError) newErrors.title = titleError;
    if (descriptionError) newErrors.description = descriptionError;

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = () => {
    if (!validateForm()) return;

    const data: TodoCreateInput | TodoUpdateInput = {
      title: title.trim(),
      description: description.trim() || undefined,
      priority,
      assigned_to: assignedTo || undefined,
      due_date: dueDate || undefined,
    };

    if (isEditing) {
      (data as TodoUpdateInput).status = status;
    }

    onSubmit(data);
  };

  return (
    <ScrollView style={styles.container} showsVerticalScrollIndicator={false}>
      <View style={styles.section}>
        <Text style={styles.label}>Title *</Text>
        <TextInput
          style={[styles.input, errors.title && styles.inputError]}
          placeholder="Enter todo title"
          value={title}
          onChangeText={(text) => {
            setTitle(text);
            if (errors.title) {
              setErrors((prev) => ({ ...prev, title: undefined }));
            }
          }}
          maxLength={200}
          editable={!isLoading}
        />
        {errors.title && <Text style={styles.errorText}>{errors.title}</Text>}
      </View>

      <View style={styles.section}>
        <Text style={styles.label}>Description</Text>
        <TextInput
          style={[styles.textArea, errors.description && styles.inputError]}
          placeholder="Add a description (optional)"
          value={description}
          onChangeText={(text) => {
            setDescription(text);
            if (errors.description) {
              setErrors((prev) => ({ ...prev, description: undefined }));
            }
          }}
          multiline
          numberOfLines={4}
          textAlignVertical="top"
          maxLength={2000}
          editable={!isLoading}
        />
        {errors.description && (
          <Text style={styles.errorText}>{errors.description}</Text>
        )}
        <Text style={styles.characterCount}>{description.length}/2000</Text>
      </View>

      <View style={styles.section}>
        <Text style={styles.label}>Priority</Text>
        <View style={styles.priorityContainer}>
          {priorities.map((p) => (
            <TouchableOpacity
              key={p}
              style={[
                styles.priorityButton,
                priority === p && {
                  backgroundColor: `${priorityColors[p]}20`,
                  borderColor: priorityColors[p],
                },
              ]}
              onPress={() => setPriority(p)}
              disabled={isLoading}
            >
              <View
                style={[
                  styles.priorityDot,
                  { backgroundColor: priorityColors[p] },
                ]}
              />
              <Text
                style={[
                  styles.priorityButtonText,
                  priority === p && { color: priorityColors[p] },
                ]}
              >
                {priorityLabels[p]}
              </Text>
            </TouchableOpacity>
          ))}
        </View>
      </View>

      {isEditing && (
        <View style={styles.section}>
          <Text style={styles.label}>Status</Text>
          <View style={styles.statusContainer}>
            {statuses.map((s) => (
              <TouchableOpacity
                key={s}
                style={[
                  styles.statusButton,
                  status === s && styles.statusButtonActive,
                ]}
                onPress={() => setStatus(s)}
                disabled={isLoading}
              >
                <Text
                  style={[
                    styles.statusButtonText,
                    status === s && styles.statusButtonTextActive,
                  ]}
                >
                  {s.replace('_', ' ').replace(/\b\w/g, (l) => l.toUpperCase())}
                </Text>
              </TouchableOpacity>
            ))}
          </View>
        </View>
      )}

      <View style={styles.section}>
        <Text style={styles.label}>Assign To</Text>
        <View style={styles.assignContainer}>
          <TouchableOpacity
            style={[
              styles.assignButton,
              !assignedTo && styles.assignButtonActive,
            ]}
            onPress={() => setAssignedTo('')}
            disabled={isLoading}
          >
            <Text
              style={[
                styles.assignButtonText,
                !assignedTo && styles.assignButtonTextActive,
              ]}
            >
              Me
            </Text>
          </TouchableOpacity>
          {connections.map((conn) => (
            <TouchableOpacity
              key={conn.id}
              style={[
                styles.assignButton,
                assignedTo === conn.id && styles.assignButtonActive,
              ]}
              onPress={() => setAssignedTo(conn.id)}
              disabled={isLoading}
            >
              <Text
                style={[
                  styles.assignButtonText,
                  assignedTo === conn.id && styles.assignButtonTextActive,
                ]}
                numberOfLines={1}
              >
                {conn.display_name}
              </Text>
            </TouchableOpacity>
          ))}
        </View>
      </View>

      <View style={styles.section}>
        <Text style={styles.label}>Due Date</Text>
        <TextInput
          style={styles.input}
          placeholder="YYYY-MM-DD"
          value={dueDate}
          onChangeText={setDueDate}
          editable={!isLoading}
        />
      </View>

      <View style={styles.buttonContainer}>
        <TouchableOpacity
          style={[styles.button, styles.cancelButton]}
          onPress={onCancel}
          disabled={isLoading}
        >
          <Text style={styles.cancelButtonText}>Cancel</Text>
        </TouchableOpacity>
        <TouchableOpacity
          style={[styles.button, styles.submitButton, isLoading && styles.buttonDisabled]}
          onPress={handleSubmit}
          disabled={isLoading}
        >
          <Text style={styles.submitButtonText}>
            {isLoading ? 'Saving...' : isEditing ? 'Update' : 'Create'}
          </Text>
        </TouchableOpacity>
      </View>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    padding: 20,
  },
  section: {
    marginBottom: 24,
  },
  label: {
    fontSize: 14,
    fontWeight: '600',
    color: '#1a1a1a',
    marginBottom: 8,
  },
  input: {
    borderWidth: 1,
    borderColor: '#ddd',
    borderRadius: 12,
    paddingHorizontal: 16,
    paddingVertical: 14,
    fontSize: 16,
    backgroundColor: '#fafafa',
  },
  inputError: {
    borderColor: '#ef4444',
  },
  errorText: {
    fontSize: 12,
    color: '#ef4444',
    marginTop: 4,
  },
  textArea: {
    borderWidth: 1,
    borderColor: '#ddd',
    borderRadius: 12,
    paddingHorizontal: 16,
    paddingVertical: 14,
    fontSize: 16,
    backgroundColor: '#fafafa',
    height: 100,
  },
  characterCount: {
    fontSize: 12,
    color: '#999',
    textAlign: 'right',
    marginTop: 4,
  },
  priorityContainer: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 8,
  },
  priorityButton: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 16,
    paddingVertical: 10,
    borderRadius: 12,
    backgroundColor: '#f5f5f5',
    borderWidth: 1,
    borderColor: 'transparent',
    gap: 8,
  },
  priorityDot: {
    width: 8,
    height: 8,
    borderRadius: 4,
  },
  priorityButtonText: {
    fontSize: 14,
    fontWeight: '500',
    color: '#666',
  },
  statusContainer: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 8,
  },
  statusButton: {
    paddingHorizontal: 16,
    paddingVertical: 10,
    borderRadius: 12,
    backgroundColor: '#f5f5f5',
  },
  statusButtonActive: {
    backgroundColor: '#007AFF',
  },
  statusButtonText: {
    fontSize: 14,
    fontWeight: '500',
    color: '#666',
  },
  statusButtonTextActive: {
    color: '#fff',
  },
  assignContainer: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 8,
  },
  assignButton: {
    paddingHorizontal: 16,
    paddingVertical: 10,
    borderRadius: 12,
    backgroundColor: '#f5f5f5',
    maxWidth: 150,
  },
  assignButtonActive: {
    backgroundColor: '#007AFF',
  },
  assignButtonText: {
    fontSize: 14,
    fontWeight: '500',
    color: '#666',
  },
  assignButtonTextActive: {
    color: '#fff',
  },
  buttonContainer: {
    flexDirection: 'row',
    gap: 12,
    marginTop: 12,
    marginBottom: 40,
  },
  button: {
    flex: 1,
    paddingVertical: 16,
    borderRadius: 12,
    alignItems: 'center',
  },
  cancelButton: {
    backgroundColor: '#f5f5f5',
  },
  cancelButtonText: {
    fontSize: 16,
    fontWeight: '600',
    color: '#666',
  },
  submitButton: {
    backgroundColor: '#007AFF',
  },
  submitButtonText: {
    fontSize: 16,
    fontWeight: '600',
    color: '#fff',
  },
  buttonDisabled: {
    opacity: 0.6,
  },
});
