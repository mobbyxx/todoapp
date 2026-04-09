import { Model } from '@nozbe/watermelondb';
import { field, date, readonly } from '@nozbe/watermelondb/decorators';
import type { TodoStatus, TodoPriority } from '../types';

export default class Todo extends Model {
  static table = 'todos';

  @field('title') title!: string;
  @field('description') description!: string | null;
  @field('status') status!: TodoStatus;
  @field('priority') priority!: TodoPriority;
  @field('created_by') createdBy!: string;
  @field('assigned_to') assignedTo!: string | null;
  @field('due_date') dueDate!: string | null;
  @field('completed_at') completedAt!: string | null;
  @field('version') version!: number;
  @field('tags') tags!: string | null;
  @field('server_id') serverId!: string | null;
  @field('is_synced') isSynced!: boolean;
  @field('synced_at') syncedAt!: number | null;
  @readonly @date('created_at') createdAt!: number;
  @readonly @date('updated_at') updatedAt!: number;

  get tagsArray(): string[] {
    return this.tags ? JSON.parse(this.tags) : [];
  }

  setTagsArray(tags: string[]) {
    return this.update((record: any) => {
      record.tags = JSON.stringify(tags);
    });
  }

  async toggleComplete() {
    const newStatus: TodoStatus = this.status === 'completed' ? 'pending' : 'completed';
    const completedAt = newStatus === 'completed' ? new Date().toISOString() : null;
    
    await this.update((record: any) => {
      record.status = newStatus;
      record.completedAt = completedAt;
      record.isSynced = false;
    });
  }

  async markAsSynced(serverId: string) {
    await this.update((record: any) => {
      record.serverId = serverId;
      record.isSynced = true;
      record.syncedAt = Date.now();
    });
  }

  toJSON() {
    return {
      id: this.id,
      title: this.title,
      description: this.description,
      status: this.status,
      priority: this.priority,
      created_by: this.createdBy,
      assigned_to: this.assignedTo,
      due_date: this.dueDate,
      completed_at: this.completedAt,
      version: this.version,
      tags: this.tagsArray,
      server_id: this.serverId,
      is_synced: this.isSynced,
      synced_at: this.syncedAt,
      created_at: this.createdAt,
      updated_at: this.updatedAt,
    };
  }
}
