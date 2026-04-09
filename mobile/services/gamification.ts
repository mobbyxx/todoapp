import { AxiosError } from 'axios';
import { api, handleApiError } from './api';
import {
  UserProfile,
  Badge,
  UserBadge,
  Level,
  Reward,
  UserReward,
  PointsTransaction,
  ApiResponse,
} from '../types';

export async function getUserProfile(): Promise<UserProfile> {
  try {
    const response = await api.get<ApiResponse<UserProfile>>('/users/profile');
    return response.data.data;
  } catch (error) {
    const apiError = handleApiError(error as AxiosError);
    throw new Error(apiError.message);
  }
}

export async function getBadges(): Promise<Badge[]> {
  try {
    const response = await api.get<ApiResponse<Badge[]>>('/badges');
    return response.data.data;
  } catch (error) {
    const apiError = handleApiError(error as AxiosError);
    throw new Error(apiError.message);
  }
}

export async function getUserBadges(): Promise<UserBadge[]> {
  try {
    const response = await api.get<ApiResponse<UserBadge[]>>('/users/badges');
    return response.data.data;
  } catch (error) {
    const apiError = handleApiError(error as AxiosError);
    throw new Error(apiError.message);
  }
}

export async function getRewards(): Promise<Reward[]> {
  try {
    const response = await api.get<ApiResponse<Reward[]>>('/rewards');
    return response.data.data;
  } catch (error) {
    const apiError = handleApiError(error as AxiosError);
    throw new Error(apiError.message);
  }
}

export async function getUserRewards(): Promise<UserReward[]> {
  try {
    const response = await api.get<ApiResponse<UserReward[]>>('/users/rewards');
    return response.data.data;
  } catch (error) {
    const apiError = handleApiError(error as AxiosError);
    throw new Error(apiError.message);
  }
}

export async function redeemReward(rewardId: string): Promise<UserReward> {
  try {
    const response = await api.post<ApiResponse<UserReward>>(`/rewards/${rewardId}/redeem`);
    return response.data.data;
  } catch (error) {
    const apiError = handleApiError(error as AxiosError);
    throw new Error(apiError.message);
  }
}

export async function createReward(data: {
  name: string;
  description: string;
  type: 'badge' | 'points' | 'feature';
  value: number;
}): Promise<Reward> {
  try {
    const response = await api.post<ApiResponse<Reward>>('/rewards', data);
    return response.data.data;
  } catch (error) {
    const apiError = handleApiError(error as AxiosError);
    throw new Error(apiError.message);
  }
}

export async function getCurrentLevel(): Promise<Level> {
  try {
    const response = await api.get<ApiResponse<Level>>('/users/level');
    return response.data.data;
  } catch (error) {
    const apiError = handleApiError(error as AxiosError);
    throw new Error(apiError.message);
  }
}

export async function getLevelProgress(): Promise<{
  currentPoints: number;
  nextLevelPoints: number;
  progressPercentage: number;
}> {
  try {
    const response = await api.get<ApiResponse<{
      currentPoints: number;
      nextLevelPoints: number;
      progressPercentage: number;
    }>>('/users/level/progress');
    return response.data.data;
  } catch (error) {
    const apiError = handleApiError(error as AxiosError);
    throw new Error(apiError.message);
  }
}

export async function getPointsTransactions(): Promise<PointsTransaction[]> {
  try {
    const response = await api.get<ApiResponse<PointsTransaction[]>>('/users/points/transactions');
    return response.data.data;
  } catch (error) {
    const apiError = handleApiError(error as AxiosError);
    throw new Error(apiError.message);
  }
}

export interface Goal {
  id: string;
  title: string;
  description?: string;
  target_value: number;
  current_value: number;
  unit: string;
  deadline?: string;
  created_by: string;
  participants: string[];
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export async function getGoals(): Promise<Goal[]> {
  try {
    const response = await api.get<ApiResponse<Goal[]>>('/goals');
    return response.data.data;
  } catch (error) {
    const apiError = handleApiError(error as AxiosError);
    throw new Error(apiError.message);
  }
}

export async function getActiveGoals(): Promise<Goal[]> {
  try {
    const response = await api.get<ApiResponse<Goal[]>>('/goals/active');
    return response.data.data;
  } catch (error) {
    const apiError = handleApiError(error as AxiosError);
    throw new Error(apiError.message);
  }
}

export async function joinGoal(goalId: string): Promise<Goal> {
  try {
    const response = await api.post<ApiResponse<Goal>>(`/goals/${goalId}/join`);
    return response.data.data;
  } catch (error) {
    const apiError = handleApiError(error as AxiosError);
    throw new Error(apiError.message);
  }
}

export async function getCurrentStreak(): Promise<{
  currentStreak: number;
  longestStreak: number;
  lastActiveDate: string;
}> {
  try {
    const response = await api.get<ApiResponse<{
      currentStreak: number;
      longestStreak: number;
      lastActiveDate: string;
    }>>('/users/streak');
    return response.data.data;
  } catch (error) {
    const apiError = handleApiError(error as AxiosError);
    throw new Error(apiError.message);
  }
}
