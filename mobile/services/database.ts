import { Database } from '@nozbe/watermelondb';
import SQLiteAdapter from '@nozbe/watermelondb/adapters/sqlite';
import { appSchema, tableSchema } from '@nozbe/watermelondb';
import TodoModel from '../models/Todo';
import UserModel from '../models/User';
import ConnectionModel from '../models/Connection';
import BadgeModel from '../models/Badge';

export const mySchema = appSchema({
  version: 1,
  tables: [
    tableSchema({
      name: 'todos',
      columns: [
        { name: 'title', type: 'string' },
        { name: 'description', type: 'string', isOptional: true },
        { name: 'status', type: 'string' },
        { name: 'priority', type: 'string' },
        { name: 'created_by', type: 'string' },
        { name: 'assigned_to', type: 'string', isOptional: true },
        { name: 'due_date', type: 'string', isOptional: true },
        { name: 'completed_at', type: 'string', isOptional: true },
        { name: 'version', type: 'number' },
        { name: 'tags', type: 'string', isOptional: true },
        { name: 'server_id', type: 'string', isOptional: true },
        { name: 'is_synced', type: 'boolean' },
        { name: 'synced_at', type: 'number', isOptional: true },
        { name: 'created_at', type: 'number' },
        { name: 'updated_at', type: 'number' },
      ],
    }),
    tableSchema({
      name: 'users',
      columns: [
        { name: 'email', type: 'string' },
        { name: 'display_name', type: 'string' },
        { name: 'avatar_url', type: 'string', isOptional: true },
        { name: 'server_id', type: 'string', isOptional: true },
        { name: 'is_active', type: 'boolean' },
        { name: 'last_seen_at', type: 'string', isOptional: true },
        { name: 'is_synced', type: 'boolean' },
        { name: 'synced_at', type: 'number', isOptional: true },
        { name: 'created_at', type: 'number' },
        { name: 'updated_at', type: 'number' },
      ],
    }),
    tableSchema({
      name: 'connections',
      columns: [
        { name: 'user_a_id', type: 'string' },
        { name: 'user_b_id', type: 'string' },
        { name: 'status', type: 'string' },
        { name: 'requested_by', type: 'string' },
        { name: 'server_id', type: 'string', isOptional: true },
        { name: 'is_synced', type: 'boolean' },
        { name: 'synced_at', type: 'number', isOptional: true },
        { name: 'created_at', type: 'number' },
        { name: 'updated_at', type: 'number' },
      ],
    }),
    tableSchema({
      name: 'badges',
      columns: [
        { name: 'name', type: 'string' },
        { name: 'description', type: 'string' },
        { name: 'icon_url', type: 'string', isOptional: true },
        { name: 'type', type: 'string' },
        { name: 'points_value', type: 'number' },
        { name: 'server_id', type: 'string', isOptional: true },
        { name: 'is_synced', type: 'boolean' },
        { name: 'synced_at', type: 'number', isOptional: true },
        { name: 'created_at', type: 'number' },
        { name: 'updated_at', type: 'number' },
      ],
    }),
  ],
});

const adapter = new SQLiteAdapter({
  schema: mySchema,
  dbName: 'todoapp',
  jsi: false,
  onSetUpError: (error) => {
    console.error('Database setup error:', error);
  },
});

export const database = new Database({
  adapter,
  modelClasses: [TodoModel, UserModel, ConnectionModel, BadgeModel],
});

export default database;
