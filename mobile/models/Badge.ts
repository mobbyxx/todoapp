import { Model } from '@nozbe/watermelondb';
import { field, date, readonly } from '@nozbe/watermelondb/decorators';
import type { BadgeType } from '../types';

export default class Badge extends Model {
  static table = 'badges';

  @field('name') name!: string;
  @field('description') description!: string;
  @field('icon_url') iconUrl!: string | null;
  @field('type') type!: BadgeType;
  @field('points_value') pointsValue!: number;
  @field('server_id') serverId!: string | null;
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
      name: this.name,
      description: this.description,
      icon_url: this.iconUrl,
      type: this.type,
      points_value: this.pointsValue,
      server_id: this.serverId,
      is_synced: this.isSynced,
      synced_at: this.syncedAt,
      created_at: this.createdAt,
      updated_at: this.updatedAt,
    };
  }
}
