import { useState, useEffect, useCallback } from 'react';
import { Q } from '@nozbe/watermelondb';
import { database } from '../services/database';
import Todo from '../models/Todo';
import type { TodoStatus, TodoPriority } from '../types';

export function useTodos(status?: TodoStatus | 'all') {
  const [todos, setTodos] = useState<Todo[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    setIsLoading(true);
    
    const todosCollection = database.get<Todo>('todos');
    let query = todosCollection.query();
    
    if (status && status !== 'all') {
      query = todosCollection.query(Q.where('status', status));
    }

    const subscription = query.observe().subscribe(
      (data) => {
        setTodos(data);
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
    const todosCollection = database.get<Todo>('todos');
    const createdTodo = await database.write(async () => {
      return await todosCollection.create((todo) => {
        todo.title = data.title;
        todo.description = data.description || null;
        todo.status = 'pending' as TodoStatus;
        todo.priority = (data.priority || 'medium') as TodoPriority;
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
    return createdTodo;
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
    const todo = await database.get<Todo>('todos').find(id);
    
    await database.write(async () => {
      await todo.update((record) => {
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
    const todo = await database.get<Todo>('todos').find(id);
    
    await database.write(async () => {
      await todo.markAsDeleted();
    });
  }, []);

  const toggleTodoComplete = useCallback(async (id: string) => {
    const todo = await database.get<Todo>('todos').find(id);
    
    await database.write(async () => {
      await todo.toggleComplete();
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
  const [todo, setTodo] = useState<Todo | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    if (!id) {
      setTodo(null);
      setIsLoading(false);
      return;
    }

    setIsLoading(true);

    const subscription = database
      .get<Todo>('todos')
      .findAndObserve(id)
      .subscribe(
        (data) => {
          setTodo(data);
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
