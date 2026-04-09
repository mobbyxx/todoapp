import { useEffect, useRef } from 'react';
import { AppState, AppStateStatus } from 'react-native';
import { useOfflineStore } from '../stores/offlineStore';

export function useAppSync() {
  const appState = useRef<AppStateStatus>(AppState.currentState);
  const { isOnline, initSyncListener, triggerSync } = useOfflineStore();

  useEffect(() => {
    const unsubscribe = initSyncListener();
    return unsubscribe;
  }, [initSyncListener]);

  useEffect(() => {
    const subscription = AppState.addEventListener('change', (nextAppState) => {
      if (
        appState.current.match(/inactive|background/) &&
        nextAppState === 'active'
      ) {
        if (isOnline) {
          triggerSync();
        }
      }
      appState.current = nextAppState;
    });

    return () => {
      subscription.remove();
    };
  }, [isOnline, triggerSync]);

  useEffect(() => {
    const interval = setInterval(() => {
      if (isOnline) {
        triggerSync();
      }
    }, 60000);

    return () => clearInterval(interval);
  }, [isOnline, triggerSync]);
}
