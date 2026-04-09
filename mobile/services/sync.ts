import { synchronize, SyncPullResult } from '@nozbe/watermelondb/sync';
import { AxiosError } from 'axios';
import { api, handleApiError } from './api';
import { database } from './database';
import type { Todo } from '../types';

export interface SyncStatus {
  isSyncing: boolean;
  lastSyncedAt: number | null;
  pendingChanges: number;
  isOnline: boolean;
  error: string | null;
}

let syncStatus: SyncStatus = {
  isSyncing: false,
  lastSyncedAt: null,
  pendingChanges: 0,
  isOnline: true,
  error: null,
};

const statusListeners: Set<(status: SyncStatus) => void> = new Set();

function updateStatus(updates: Partial<SyncStatus>) {
  syncStatus = { ...syncStatus, ...updates };
  statusListeners.forEach((listener) => listener(syncStatus));
}

export function subscribeToSyncStatus(listener: (status: SyncStatus) => void) {
  statusListeners.add(listener);
  listener(syncStatus);
  return () => statusListeners.delete(listener);
}

export function getSyncStatus(): SyncStatus {
  return { ...syncStatus };
}

export async function updatePendingChangesCount(): Promise<number> {
  try {
    const todosCollection = database.get('todos');
    const allTodos = await todosCollection.query().fetch();
    
    const pendingCount = allTodos.filter((todo) => {
      const record = todo._raw;
      return record._status !== 'synced';
    }).length;
    updateStatus({ pendingChanges: pendingCount });
    return pendingCount;
  } catch (error) {
    console.error('Failed to count pending changes:', error);
    return 0;
  }
}

export async function syncDatabase(): Promise<void> {
  if (syncStatus.isSyncing) {
    return;
  }

  updateStatus({ isSyncing: true, error: null });

  try {
    await synchronize({
      database,
      pullChanges: async ({ lastPulledAt }) => {
        try {
          const response = await api.get('/sync/todos', {
            params: { last_synced_at: lastPulledAt },
          });

          const { changes, timestamp } = response.data.data;

          return {
            changes: {
              todos: {
                created: changes.todos?.created?.map((todo: Todo) => ({
                  ...todo,
                  server_id: todo.id,
                  id: `server_${todo.id}`,
                  is_synced: true,
                  synced_at: timestamp,
                  created_at: new Date(todo.created_at).getTime(),
                  updated_at: new Date(todo.updated_at).getTime(),
                  tags: todo.tags ? JSON.stringify(todo.tags) : null,
                })) || [],
                updated: changes.todos?.updated?.map((todo: Todo) => ({
                  ...todo,
                  server_id: todo.id,
                  id: `server_${todo.id}`,
                  is_synced: true,
                  synced_at: timestamp,
                  created_at: new Date(todo.created_at).getTime(),
                  updated_at: new Date(todo.updated_at).getTime(),
                  tags: todo.tags ? JSON.stringify(todo.tags) : null,
                })) || [],
                deleted: changes.todos?.deleted?.map((id: string) => `server_${id}`) || [],
              },
            },
            timestamp,
          } as SyncPullResult;
        } catch (error) {
          const apiError = handleApiError(error as AxiosError);
          throw new Error(`Pull failed: ${apiError.message}`);
        }
      },
      pushChanges: async ({ changes, lastPulledAt }) => {
        try {
          const todosToPush = {
            created: changes.todos?.created?.map((todo: any) => ({
              title: todo.title,
              description: todo.description,
              status: todo.status,
              priority: todo.priority,
              created_by: todo.created_by,
              assigned_to: todo.assigned_to,
              due_date: todo.due_date,
              completed_at: todo.completed_at,
              version: todo.version || 1,
              tags: todo.tags ? JSON.parse(todo.tags) : [],
            })) || [],
            updated: changes.todos?.updated?.map((todo: any) => ({
              id: todo.server_id || todo.id,
              title: todo.title,
              description: todo.description,
              status: todo.status,
              priority: todo.priority,
              assigned_to: todo.assigned_to,
              due_date: todo.due_date,
              completed_at: todo.completed_at,
              version: todo.version,
              tags: todo.tags ? JSON.parse(todo.tags) : [],
            })) || [],
            deleted: changes.todos?.deleted || [],
          };

          if (
            todosToPush.created.length > 0 ||
            todosToPush.updated.length > 0 ||
            todosToPush.deleted.length > 0
          ) {
            await api.post('/sync/todos', {
              changes: { todos: todosToPush },
              last_synced_at: lastPulledAt,
            });
          }
        } catch (error) {
          const apiError = handleApiError(error as AxiosError);
          throw new Error(`Push failed: ${apiError.message}`);
        }
      },
      migrationsEnabledAtVersion: 1,
    });

    await updatePendingChangesCount();
    updateStatus({
      lastSyncedAt: Date.now(),
      isOnline: true,
      isSyncing: false,
    });
  } catch (error) {
    const errorMessage = error instanceof Error ? error.message : 'Sync failed';
    updateStatus({
      error: errorMessage,
      isSyncing: false,
      isOnline: false,
    });
    throw error;
  }
}

export async function checkConnectivity(): Promise<boolean> {
  try {
    await api.get('/health', { timeout: 5000 });
    updateStatus({ isOnline: true });
    return true;
  } catch {
    updateStatus({ isOnline: false });
    return false;
  }
}

export async function resolveConflict(
  localTodo: Todo,
  remoteTodo: Todo,
  strategy: 'local' | 'remote' | 'merge' = 'remote'
): Promise<Todo> {
  switch (strategy) {
    case 'local':
      return localTodo;
    case 'remote':
      return remoteTodo;
    case 'merge':
      return {
        ...remoteTodo,
        status: localTodo.status,
        completed_at: localTodo.completed_at,
      };
    default:
      return remoteTodo;
  }
}

export function useSyncStatus() {
  return {
    getStatus: getSyncStatus,
    subscribe: subscribeToSyncStatus,
    sync: syncDatabase,
    checkConnectivity,
    updatePendingChangesCount,
  };
}
