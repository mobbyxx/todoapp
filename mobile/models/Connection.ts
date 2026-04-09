import { Model } from '@nozbe/watermelondb';
import { field, date, readonly } from '@nozbe/watermelondb/decorators';
import type { ConnectionStatus } from '../types';

export default class Connection extends Model {
  static table = 'connections';

  @field('user_a_id') userAId!: string;
  @field('user_b_id') userBId!: string;
  @field('status') status!: ConnectionStatus;
  @field('requested_by') requestedBy!: string;
  @field('server_id') serverId!: string | null;
  @field('is_synced') isSynced!: boolean;
  @field('synced_at') syncedAt!: number | null;
  @readonly @date('created_at') createdAt!: number;
  @readonly @date('updated_at') updatedAt!: number;

  async updateStatus(newStatus: ConnectionStatus) {
    await this.update((record: any) => {
      record.status = newStatus;
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
      user_a_id: this.userAId,
      user_b_id: this.userBId,
      status: this.status,
      requested_by: this.requestedBy,
      server_id: this.serverId,
      is_synced: this.isSynced,
      synced_at: this.syncedAt,
      created_at: this.createdAt,
      updated_at: this.updatedAt,
    };
  }
}
