import { Model } from '@nozbe/watermelondb';
import { field, date, readonly } from '@nozbe/watermelondb/decorators';

export default class User extends Model {
  static table = 'users';

  @field('email') email!: string;
  @field('display_name') displayName!: string;
  @field('avatar_url') avatarUrl!: string | null;
  @field('server_id') serverId!: string | null;
  @field('is_active') isActive!: boolean;
  @field('last_seen_at') lastSeenAt!: string | null;
  @field('is_synced') isSynced!: boolean;
  @field('synced_at') syncedAt!: number | null;
  @readonly @date('created_at') createdAt!: number;
  @readonly @date('updated_at') updatedAt!: number;

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
      email: this.email,
      display_name: this.displayName,
      avatar_url: this.avatarUrl,
      server_id: this.serverId,
      is_active: this.isActive,
      last_seen_at: this.lastSeenAt,
      is_synced: this.isSynced,
      synced_at: this.syncedAt,
      created_at: this.createdAt,
      updated_at: this.updatedAt,
    };
  }
}
