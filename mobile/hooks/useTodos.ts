import { useState, useEffect, useCallback } from 'react';
import { Q } from '@nozbe/watermelondb';
import { database } from '../services/database';
import type { TodoStatus, TodoPriority } from '../types';

interface TodoModel {
  id: string;
  title: string;
  description: string | null;
  status: TodoStatus;
  priority: TodoPriority;
  createdBy: string;
  assignedTo: string | null;
  dueDate: string | null;
  completedAt: string | null;
  version: number;
  tags: string | null;
  serverId: string | null;
  isSynced: boolean;
  syncedAt: number | null;
  createdAt: number;
  updatedAt: number;
  toggleComplete(): Promise<void>;
  markAsDeleted(): Promise<void>;
  update(updater: (record: any) => void): Promise<void>;
  toJSON(): any;
}

export function useTodos(status?: TodoStatus | 'all') {
  const [todos, setTodos] = useState<TodoModel[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    setIsLoading(true);
    
    const todosCollection = database.get('todos');
    let query = todosCollection.query();
    
    if (status && status !== 'all') {
      query = todosCollection.query(Q.where('status', status));
    }

    const subscription = query.observe().subscribe(
      (data) => {
        setTodos(data as unknown as TodoModel[]);
        setIsLoading(false);
      },
      (error) => {
        console.error('Failed to observe todos:', error);
        setIsLoading(false);
      }
    );

    return () => subscription.unsubscribe();
  }, [status]);

  const createTodo = useCallback(async (data: {
    title: string;
    description?: string;
    priority?: TodoPriority;
    dueDate?: string;
    assignedTo?: string;
    tags?: string[];
  }) => {
    const todosCollection = database.get('todos');
    const createdTodo = await database.write(async () => {
      return await todosCollection.create((todo: any) => {
        todo.title = data.title;
        todo.description = data.description || null;
        todo.status = 'pending';
        todo.priority = data.priority || 'medium';
        todo.createdBy = 'current_user';
        todo.assignedTo = data.assignedTo || null;
        todo.dueDate = data.dueDate || null;
        todo.completedAt = null;
        todo.version = 1;
        todo.tags = data.tags ? JSON.stringify(data.tags) : null;
        todo.serverId = null;
        todo.isSynced = false;
        todo.syncedAt = null;
      });
    });
    return createdTodo as unknown as TodoModel;
  }, []);

  const updateTodo = useCallback(async (id: string, updates: {
    title?: string;
    description?: string;
    status?: TodoStatus;
    priority?: TodoPriority;
    dueDate?: string;
    assignedTo?: string;
    tags?: string[];
  }) => {
    const todo = await database.get('todos').find(id) as unknown as TodoModel;
    
    await database.write(async () => {
      await todo.update((record: any) => {
        if (updates.title !== undefined) record.title = updates.title;
        if (updates.description !== undefined) record.description = updates.description || null;
        if (updates.status !== undefined) record.status = updates.status;
        if (updates.priority !== undefined) record.priority = updates.priority;
        if (updates.dueDate !== undefined) record.dueDate = updates.dueDate || null;
        if (updates.assignedTo !== undefined) record.assignedTo = updates.assignedTo || null;
        if (updates.tags !== undefined) record.tags = updates.tags ? JSON.stringify(updates.tags) : null;
        record.isSynced = false;
        record.version = (record.version || 0) + 1;
      });
    });
    
    return todo;
  }, []);

  const deleteTodo = useCallback(async (id: string) => {
    const todo = await database.get('todos').find(id) as unknown as TodoModel;
    
    await database.write(async () => {
      await todo.markAsDeleted();
    });
  }, []);

  const toggleTodoComplete = useCallback(async (id: string) => {
    const todo = await database.get('todos').find(id) as unknown as TodoModel;
    const newStatus: TodoStatus = todo.status === 'completed' ? 'pending' : 'completed';
    const completedAt = newStatus === 'completed' ? new Date().toISOString() : null;
    
    await database.write(async () => {
      await todo.update((record: any) => {
        record.status = newStatus;
        record.completedAt = completedAt;
        record.isSynced = false;
      });
    });
    
    return todo;
  }, []);

  return {
    todos,
    isLoading,
    createTodo,
    updateTodo,
    deleteTodo,
    toggleTodoComplete,
  };
}

export function useTodo(id: string | null) {
  const [todo, setTodo] = useState<TodoModel | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    if (!id) {
      setTodo(null);
      setIsLoading(false);
      return;
    }

    setIsLoading(true);

    const subscription = database
      .get('todos')
      .findAndObserve(id)
      .subscribe(
        (data) => {
          setTodo(data as unknown as TodoModel);
          setIsLoading(false);
        },
        (error) => {
          console.error('Failed to observe todo:', error);
          setTodo(null);
          setIsLoading(false);
        }
      );

    return () => subscription.unsubscribe();
  }, [id]);

  return { todo, isLoading };
}
