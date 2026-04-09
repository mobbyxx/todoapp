export type TodoStatus = 'pending' | 'in_progress' | 'completed' | 'archived';

export type TodoPriority = 'low' | 'medium' | 'high' | 'urgent';

export type ConnectionStatus = 'pending' | 'accepted' | 'blocked';

export type BadgeType = 'achievement' | 'milestone' | 'special';

export type RewardType = 'badge' | 'points' | 'feature';

export interface User {
  id: string;
  email: string;
  display_name: string;
  avatar_url?: string;
  created_at: string;
  updated_at: string;
  last_seen_at?: string;
  is_active: boolean;
}

export interface UserProfile extends User {
  total_todos: number;
  completed_todos: number;
  current_level: Level;
  total_points: number;
  badges: Badge[];
}

export interface Todo {
  id: string;
  title: string;
  description?: string;
  status: TodoStatus;
  priority: TodoPriority;
  created_by: string;
  assigned_to?: string;
  due_date?: string;
  completed_at?: string;
  version: number;
  created_at: string;
  updated_at: string;
  tags?: string[];
}

export interface TodoCreateInput {
  title: string;
  description?: string;
  priority?: TodoPriority;
  assigned_to?: string;
  due_date?: string;
  tags?: string[];
}

export interface TodoUpdateInput {
  title?: string;
  description?: string;
  status?: TodoStatus;
  priority?: TodoPriority;
  assigned_to?: string;
  due_date?: string;
  tags?: string[];
}

export interface Connection {
  id: string;
  user_a_id: string;
  user_b_id: string;
  status: ConnectionStatus;
  created_at: string;
  updated_at: string;
  requested_by: string;
}

export interface ConnectionWithUsers extends Connection {
  user_a: User;
  user_b: User;
}

export interface Badge {
  id: string;
  name: string;
  description: string;
  icon_url: string;
  type: BadgeType;
  points_value: number;
  created_at: string;
}

export interface UserBadge {
  id: string;
  user_id: string;
  badge_id: string;
  badge: Badge;
  awarded_at: string;
  awarded_by?: string;
}

export interface Level {
  id: string;
  level_number: number;
  name: string;
  min_points: number;
  max_points: number;
  icon_url?: string;
}

export interface Reward {
  id: string;
  name: string;
  description: string;
  type: RewardType;
  value: number;
  icon_url?: string;
  created_at: string;
}

export interface UserReward {
  id: string;
  user_id: string;
  reward_id: string;
  reward: Reward;
  claimed_at: string;
  expires_at?: string;
}

export interface PointsTransaction {
  id: string;
  user_id: string;
  amount: number;
  reason: string;
  reference_type?: string;
  reference_id?: string;
  created_at: string;
}

export interface ApiResponse<T> {
  data: T;
  success: boolean;
  message?: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  pageSize: number;
  hasMore: boolean;
}

export interface ApiError {
  code: string;
  message: string;
  details?: Record<string, unknown>;
}

export interface SyncPayload<T> {
  items: T[];
  lastSyncedAt: string;
  clientVersion?: number;
}

export interface SyncResult<T> {
  items: T[];
  serverVersion: number;
  conflicts?: Array<{
    local: T;
    remote: T;
  }>;
}

export interface Notification {
  id: string;
  user_id: string;
  type: 'todo_assigned' | 'todo_completed' | 'connection_request' | 'connection_accepted' | 'badge_awarded' | 'level_up';
  title: string;
  body: string;
  data?: Record<string, unknown>;
  is_read: boolean;
  created_at: string;
}

export interface DeepLinkData {
  path: string;
  params: Record<string, string>;
}

export type DeepLinkPath = 
  | '/todos/:id'
  | '/connections'
  | '/profile'
  | '/rewards'
  | '/settings';
