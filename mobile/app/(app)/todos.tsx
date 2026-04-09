import { useState, useCallback, useMemo } from 'react';
import {
  View,
  Text,
  StyleSheet,
  TouchableOpacity,
  Alert,
  RefreshControl,
} from 'react-native';
import { Stack, router } from 'expo-router';
import { useTodos } from '../../hooks/useTodos';
import { useOfflineStore } from '../../stores/offlineStore';
import { syncDatabase } from '../../services/sync';
import { TodoFilter } from '../../components/TodoFilter';
import { OfflineIndicator } from '../../components/OfflineIndicator';
import { FlashList, ListRenderItem } from '@shopify/flash-list';
import { TodoItem } from '../../components/TodoItem';
import type { TodoStatus } from '../../types';
import Todo from '../../models/Todo';

const ESTIMATED_ITEM_HEIGHT = 100;

type TodoItemData = Todo;

export default function TodosScreen() {
  const [selectedFilter, setSelectedFilter] = useState<TodoStatus | 'all'>('all');
  const [isRefreshing, setIsRefreshing] = useState(false);
  const { isOnline } = useOfflineStore();
  
  const { todos, isLoading, toggleTodoComplete } = useTodos(
    selectedFilter === 'all' ? undefined : selectedFilter
  );

  const filteredTodos = useMemo(() => {
    if (selectedFilter === 'all') return todos;
    return todos.filter((todo) => todo.status === selectedFilter);
  }, [todos, selectedFilter]);

  const filterCounts = useMemo(() => {
    const all = todos.length;
    const pending = todos.filter((t) => t.status === 'pending').length;
    const in_progress = todos.filter((t) => t.status === 'in_progress').length;
    const completed = todos.filter((t) => t.status === 'completed').length;
    return { all, pending, in_progress, completed };
  }, [todos]);

  const handleRefresh = useCallback(async () => {
    setIsRefreshing(true);
    try {
      if (isOnline) {
        await syncDatabase();
      }
    } catch (error) {
      console.error('Refresh failed:', error);
    } finally {
      setIsRefreshing(false);
    }
  }, [isOnline]);

  const handleTodoPress = useCallback((id: string) => {
    router.push(`/(app)/todos/${id}`);
  }, []);

  const handleToggleComplete = useCallback(
    async (id: string) => {
      try {
        await toggleTodoComplete(id);
      } catch (error) {
        Alert.alert('Error', 'Failed to update todo status');
      }
    },
    [toggleTodoComplete]
  );

  const handleFilterChange = useCallback(
    (status: TodoStatus | 'all') => {
      setSelectedFilter(status);
    },
    []
  );

  const renderItem: ListRenderItem<TodoItemData> = useCallback(
    ({ item }) => (
      <TodoItem
        todo={{
          id: item.id,
          title: item.title,
          description: item.description || undefined,
          status: item.status,
          priority: item.priority,
          due_date: item.dueDate || undefined,
        }}
        onPress={() => handleTodoPress(item.id)}
        onToggleComplete={() => handleToggleComplete(item.id)}
      />
    ),
    [handleTodoPress, handleToggleComplete]
  );

  const keyExtractor = (item: TodoItemData) => item.id;

  const renderEmptyComponent = () => (
    <View style={styles.emptyContainer}>
      <Text style={styles.emptyEmoji}>📝</Text>
      <Text style={styles.emptyTitle}>No Todos</Text>
      <Text style={styles.emptyMessage}>No todos yet. Create one to get started!</Text>
    </View>
  );

  return (
    <View style={styles.container}>
      <OfflineIndicator />
      
      <Stack.Screen
        options={{
          title: 'My Todos',
          headerRight: () => (
            <TouchableOpacity
              style={styles.addButton}
              onPress={() => router.push('/(app)/todos/create')}
            >
              <View style={styles.addButtonCircle}>
                <View style={styles.addIcon}>
                  <View style={styles.addIconHorizontal} />
                  <View style={styles.addIconVertical} />
                </View>
              </View>
            </TouchableOpacity>
          ),
        }}
      />

      <TodoFilter
        selectedFilter={selectedFilter}
        onFilterChange={handleFilterChange}
        counts={filterCounts}
      />

      {isLoading && todos.length === 0 ? (
        <View style={styles.loadingContainer}>
          <Text style={styles.loadingText}>Loading todos...</Text>
        </View>
      ) : (
        <FlashList
          data={filteredTodos}
          renderItem={renderItem}
          keyExtractor={keyExtractor}
          estimatedItemSize={ESTIMATED_ITEM_HEIGHT}
          contentContainerStyle={styles.listContent}
          showsVerticalScrollIndicator={false}
          refreshControl={
            <RefreshControl
              refreshing={isRefreshing}
              onRefresh={handleRefresh}
              tintColor="#007AFF"
            />
          }
          ListEmptyComponent={renderEmptyComponent}
          ItemSeparatorComponent={() => <View style={styles.separator} />}
        />
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f5f5f5',
    paddingTop: 28,
  },
  addButton: {
    marginRight: 8,
    padding: 8,
  },
  addButtonCircle: {
    width: 32,
    height: 32,
    borderRadius: 16,
    backgroundColor: '#007AFF',
    alignItems: 'center',
    justifyContent: 'center',
  },
  addIcon: {
    width: 14,
    height: 14,
    alignItems: 'center',
    justifyContent: 'center',
  },
  addIconHorizontal: {
    position: 'absolute',
    width: 14,
    height: 2,
    backgroundColor: '#fff',
    borderRadius: 1,
  },
  addIconVertical: {
    position: 'absolute',
    width: 2,
    height: 14,
    backgroundColor: '#fff',
    borderRadius: 1,
  },
  listContent: {
    paddingVertical: 8,
  },
  separator: {
    height: 0,
  },
  loadingContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  loadingText: {
    fontSize: 16,
    color: '#666',
  },
  emptyContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    paddingVertical: 60,
    paddingHorizontal: 40,
  },
  emptyEmoji: {
    fontSize: 48,
    marginBottom: 16,
  },
  emptyTitle: {
    fontSize: 20,
    fontWeight: '600',
    color: '#1a1a1a',
    marginBottom: 8,
  },
  emptyMessage: {
    fontSize: 14,
    color: '#666',
    textAlign: 'center',
  },
});
