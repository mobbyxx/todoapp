import { create } from 'zustand';
import { Todo, TodoStatus, TodoPriority } from '../types';

interface TodoFilters {
  status: TodoStatus | 'all';
}

interface TodoState {
  todos: Todo[];
  selectedTodo: Todo | null;
  filters: TodoFilters;
  isLoading: boolean;
  error: string | null;
}

interface TodoActions {
  setTodos: (todos: Todo[]) => void;
  addTodo: (todo: Todo) => void;
  updateTodo: (id: string, updates: Partial<Todo>) => void;
  removeTodo: (id: string) => void;
  setSelectedTodo: (todo: Todo | null) => void;
  setFilters: (filters: Partial<TodoFilters>) => void;
  setLoading: (isLoading: boolean) => void;
  setError: (error: string | null) => void;
  clearError: () => void;
}

const initialState: TodoState = {
  todos: [],
  selectedTodo: null,
  filters: {
    status: 'all',
  },
  isLoading: false,
  error: null,
};

export const useTodoStore = create<TodoState & TodoActions>()((set) => ({
  ...initialState,

  setTodos: (todos) =>
    set((state) => ({
      ...state,
      todos,
    })),

  addTodo: (todo) =>
    set((state) => ({
      ...state,
      todos: [todo, ...state.todos],
    })),

  updateTodo: (id, updates) =>
    set((state) => ({
      ...state,
      todos: state.todos.map((todo) =>
        todo.id === id ? { ...todo, ...updates } : todo
      ),
    })),

  removeTodo: (id) =>
    set((state) => ({
      ...state,
      todos: state.todos.filter((todo) => todo.id !== id),
    })),

  setSelectedTodo: (todo) =>
    set((state) => ({
      ...state,
      selectedTodo: todo,
    })),

  setFilters: (filters) =>
    set((state) => ({
      ...state,
      filters: { ...state.filters, ...filters },
    })),

  setLoading: (isLoading) =>
    set((state) => ({
      ...state,
      isLoading,
    })),

  setError: (error) =>
    set((state) => ({
      ...state,
      error,
    })),

  clearError: () =>
    set((state) => ({
      ...state,
      error: null,
    })),
}));

export const priorityColors: Record<TodoPriority, string> = {
  low: '#22c55e',
  medium: '#eab308',
  high: '#f97316',
  urgent: '#ef4444',
};

export const priorityLabels: Record<TodoPriority, string> = {
  low: 'Low',
  medium: 'Medium',
  high: 'High',
  urgent: 'Urgent',
};

export const statusLabels: Record<TodoStatus, string> = {
  pending: 'Pending',
  in_progress: 'In Progress',
  completed: 'Completed',
  archived: 'Archived',
};
