import { useEffect, useState } from 'react';
import { View, Text, StyleSheet, Animated } from 'react-native';
import { useOfflineStore } from '../stores/offlineStore';

export function OfflineIndicator() {
  const [visible, setVisible] = useState(false);
  const { isOnline, isSyncing, pendingChanges } = useOfflineStore();
  const translateY = useState(new Animated.Value(-50))[0];

  useEffect(() => {
    if (!isOnline || isSyncing || pendingChanges > 0) {
      setVisible(true);
      Animated.spring(translateY, {
        toValue: 0,
        useNativeDriver: true,
      }).start();
    } else {
      Animated.timing(translateY, {
        toValue: -50,
        duration: 300,
        useNativeDriver: true,
      }).start(() => setVisible(false));
    }
  }, [isOnline, isSyncing, pendingChanges]);

  if (!visible) return null;

  const backgroundColor = !isOnline ? '#ef4444' : isSyncing ? '#3b82f6' : '#22c55e';
  const message = !isOnline 
    ? 'Offline - Changes saved locally'
    : isSyncing 
      ? `Syncing...${pendingChanges > 0 ? ` (${pendingChanges} pending)` : ''}`
      : `${pendingChanges} changes pending sync`;

  return (
    <Animated.View style={[styles.container, { backgroundColor, transform: [{ translateY }] }]}>
      <View style={styles.content}>
        <View style={[styles.dot, { backgroundColor: '#fff' }]} />
        <Text style={styles.text}>{message}</Text>
      </View>
    </Animated.View>
  );
}

const styles = StyleSheet.create({
  container: {
    position: 'absolute',
    top: 0,
    left: 0,
    right: 0,
    zIndex: 1000,
  },
  content: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    paddingVertical: 8,
    paddingHorizontal: 16,
  },
  dot: {
    width: 8,
    height: 8,
    borderRadius: 4,
    marginRight: 8,
  },
  text: {
    color: '#fff',
    fontSize: 14,
    fontWeight: '500',
  },
});
