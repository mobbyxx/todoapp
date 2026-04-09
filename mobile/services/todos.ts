import { AxiosError } from 'axios';
import { api, handleApiError } from './api';
import {
  Todo,
  TodoCreateInput,
  TodoUpdateInput,
  TodoStatus,
  ApiResponse,
  PaginatedResponse,
} from '../types';

interface GetTodosParams {
  status?: TodoStatus | 'all';
  page?: number;
  pageSize?: number;
}

/**
 * Get all todos with optional filtering
 */
export async function getTodos(
  params: GetTodosParams = {}
): Promise<PaginatedResponse<Todo>> {
  try {
    const { status, page = 1, pageSize = 50 } = params;
    const queryParams = new URLSearchParams();

    if (status && status !== 'all') {
      queryParams.append('status', status);
    }
    queryParams.append('page', page.toString());
    queryParams.append('page_size', pageSize.toString());

    const response = await api.get<ApiResponse<PaginatedResponse<Todo>>>(
      `/todos?${queryParams.toString()}`
    );

    return response.data.data;
  } catch (error) {
    const apiError = handleApiError(error as AxiosError);
    throw new Error(apiError.message);
  }
}

/**
 * Get a single todo by ID
 */
export async function getTodo(id: string): Promise<Todo> {
  try {
    const response = await api.get<ApiResponse<Todo>>(`/todos/${id}`);
    return response.data.data;
  } catch (error) {
    const apiError = handleApiError(error as AxiosError);
    throw new Error(apiError.message);
  }
}

/**
 * Create a new todo
 */
export async function createTodo(data: TodoCreateInput): Promise<Todo> {
  try {
    const response = await api.post<ApiResponse<Todo>>('/todos', data);
    return response.data.data;
  } catch (error) {
    const apiError = handleApiError(error as AxiosError);
    throw new Error(apiError.message);
  }
}

/**
 * Update an existing todo
 */
export async function updateTodo(
  id: string,
  data: TodoUpdateInput
): Promise<Todo> {
  try {
    const response = await api.put<ApiResponse<Todo>>(`/todos/${id}`, data);
    return response.data.data;
  } catch (error) {
    const apiError = handleApiError(error as AxiosError);
    throw new Error(apiError.message);
  }
}

/**
 * Delete a todo
 */
export async function deleteTodo(id: string): Promise<void> {
  try {
    await api.delete(`/todos/${id}`);
  } catch (error) {
    const apiError = handleApiError(error as AxiosError);
    throw new Error(apiError.message);
  }
}

/**
 * Toggle todo completion status
 */
export async function toggleTodoComplete(
  id: string,
  currentStatus: TodoStatus
): Promise<Todo> {
  const newStatus: TodoStatus =
    currentStatus === 'completed' ? 'pending' : 'completed';
  return updateTodo(id, { status: newStatus });
}
