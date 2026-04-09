import { View, Text, RefreshControl, StyleSheet } from 'react-native';
import { FlashList, ListRenderItem } from '@shopify/flash-list';
import { Todo } from '../types';
import { TodoItem } from './TodoItem';

interface TodoListProps {
  todos: Todo[];
  isLoading: boolean;
  isRefreshing: boolean;
  onRefresh: () => void;
  onTodoPress: (id: string) => void;
  onToggleComplete: (id: string) => void;
  emptyMessage?: string;
}

const ESTIMATED_ITEM_HEIGHT = 100;

export function TodoList({
  todos,
  isLoading,
  isRefreshing,
  onRefresh,
  onTodoPress,
  onToggleComplete,
  emptyMessage = 'No todos yet. Create one to get started!',
}: TodoListProps) {
  const renderItem: ListRenderItem<Todo> = ({ item }) => (
    <TodoItem
      todo={item}
      onPress={() => onTodoPress(item.id)}
      onToggleComplete={() => onToggleComplete(item.id)}
    />
  );

  const keyExtractor = (item: Todo) => item.id;

  const renderEmptyComponent = () => (
    <View style={styles.emptyContainer}>
      <Text style={styles.emptyEmoji}>📝</Text>
      <Text style={styles.emptyTitle}>No Todos</Text>
      <Text style={styles.emptyMessage}>{emptyMessage}</Text>
    </View>
  );

  if (isLoading && todos.length === 0) {
    return (
      <View style={styles.loadingContainer}>
        <Text style={styles.loadingText}>Loading todos...</Text>
      </View>
    );
  }

  return (
    <FlashList
      data={todos}
      renderItem={renderItem}
      keyExtractor={keyExtractor}
      estimatedItemSize={ESTIMATED_ITEM_HEIGHT}
      contentContainerStyle={styles.listContent}
      showsVerticalScrollIndicator={false}
      refreshControl={
        <RefreshControl
          refreshing={isRefreshing}
          onRefresh={onRefresh}
          tintColor="#007AFF"
        />
      }
      ListEmptyComponent={renderEmptyComponent}
      ItemSeparatorComponent={() => <View style={styles.separator} />}
    />
  );
}

const styles = StyleSheet.create({
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
