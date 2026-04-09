import { create } from 'zustand';
import { subscribeToSyncStatus, SyncStatus, checkConnectivity, syncDatabase } from '../services/sync';

interface OfflineState extends SyncStatus {
  showOfflineToast: boolean;
  lastErrorToast: string | null;
}

interface OfflineActions {
  setIsOnline: (isOnline: boolean) => void;
  setIsSyncing: (isSyncing: boolean) => void;
  setLastSyncedAt: (timestamp: number | null) => void;
  setPendingChanges: (count: number) => void;
  setError: (error: string | null) => void;
  setShowOfflineToast: (show: boolean) => void;
  setLastErrorToast: (message: string | null) => void;
  triggerSync: () => Promise<void>;
  checkConnection: () => Promise<boolean>;
  initSyncListener: () => () => void;
}

const initialState: OfflineState = {
  isSyncing: false,
  lastSyncedAt: null,
  pendingChanges: 0,
  isOnline: true,
  error: null,
  showOfflineToast: false,
  lastErrorToast: null,
};

export const useOfflineStore = create<OfflineState & OfflineActions>()((set, get) => ({
  ...initialState,

  setIsOnline: (isOnline) => {
    const wasOffline = !get().isOnline && isOnline;
    set({ isOnline });
    if (wasOffline) {
      get().triggerSync();
    }
  },

  setIsSyncing: (isSyncing) => set({ isSyncing }),
  setLastSyncedAt: (lastSyncedAt) => set({ lastSyncedAt }),
  setPendingChanges: (pendingChanges) => set({ pendingChanges }),
  setError: (error) => set({ error }),
  setShowOfflineToast: (showOfflineToast) => set({ showOfflineToast }),
  setLastErrorToast: (lastErrorToast) => set({ lastErrorToast }),

  triggerSync: async () => {
    if (get().isSyncing || !get().isOnline) return;
    try {
      await syncDatabase();
    } catch (error) {
      console.error('Sync failed:', error);
    }
  },

  checkConnection: async () => {
    const isOnline = await checkConnectivity();
    set({ isOnline });
    return isOnline;
  },

  initSyncListener: () => {
    const unsubscribe = subscribeToSyncStatus((status) => {
      set({
        isSyncing: status.isSyncing,
        lastSyncedAt: status.lastSyncedAt,
        pendingChanges: status.pendingChanges,
        isOnline: status.isOnline,
        error: status.error,
      });
    });
    return unsubscribe;
  },
}));
