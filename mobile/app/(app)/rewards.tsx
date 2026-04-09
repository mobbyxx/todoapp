import { useState, useCallback } from 'react';
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  RefreshControl,
  TouchableOpacity,
  TextInput,
  ActivityIndicator,
  Alert,
  Modal,
} from 'react-native';
import { Stack } from 'expo-router';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  getRewards,
  getUserRewards,
  redeemReward,
  createReward,
} from '../../services/gamification';
import { RewardCard } from '../../components/gamification/RewardCard';
import { Reward, RewardType } from '../../types';

type Tab = 'available' | 'redeemed' | 'create';

export default function RewardsScreen() {
  const queryClient = useQueryClient();
  const [activeTab, setActiveTab] = useState<Tab>('available');
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [isCreateModalVisible, setIsCreateModalVisible] = useState(false);

  const [newReward, setNewReward] = useState({
    name: '',
    description: '',
    type: 'points' as RewardType,
    value: '',
  });

  const { data: allRewards, isLoading: isLoadingRewards, refetch: refetchRewards } = useQuery({
    queryKey: ['rewards'],
    queryFn: getRewards,
  });

  const { data: userRewards, isLoading: isLoadingUserRewards, refetch: refetchUserRewards } = useQuery({
    queryKey: ['userRewards'],
    queryFn: getUserRewards,
  });

  const redeemMutation = useMutation({
    mutationFn: redeemReward,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['userRewards'] });
      Alert.alert('Success', 'Reward redeemed successfully!');
    },
    onError: (error: Error) => {
      Alert.alert('Error', error.message || 'Failed to redeem reward');
    },
  });

  const createMutation = useMutation({
    mutationFn: (data: { name: string; description: string; type: RewardType; value: number }) =>
      createReward(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['rewards'] });
      setIsCreateModalVisible(false);
      setNewReward({ name: '', description: '', type: 'points', value: '' });
      Alert.alert('Success', 'Reward created successfully!');
    },
    onError: (error: Error) => {
      Alert.alert('Error', error.message || 'Failed to create reward');
    },
  });

  const isLoading = isLoadingRewards || isLoadingUserRewards;

  const handleRefresh = useCallback(async () => {
    setIsRefreshing(true);
    try {
      await Promise.all([refetchRewards(), refetchUserRewards()]);
    } finally {
      setIsRefreshing(false);
    }
  }, [refetchRewards, refetchUserRewards]);

  const handleRedeem = useCallback((reward: Reward) => {
    Alert.alert(
      'Redeem Reward',
      `Are you sure you want to redeem "${reward.name}"?`,
      [
        { text: 'Cancel', style: 'cancel' },
        { text: 'Redeem', onPress: () => redeemMutation.mutate(reward.id) },
      ]
    );
  }, [redeemMutation]);

  const handleCreateReward = useCallback(() => {
    if (!newReward.name.trim() || !newReward.description.trim() || !newReward.value) {
      Alert.alert('Error', 'Please fill in all fields');
      return;
    }

    createMutation.mutate({
      name: newReward.name,
      description: newReward.description,
      type: newReward.type,
      value: parseInt(newReward.value, 10),
    });
  }, [newReward, createMutation]);

  const rewards = allRewards || [];
  const redeemedRewards = userRewards || [];
  const redeemedRewardIds = new Set(redeemedRewards.map((ur) => ur.reward_id));

  const availableRewards = rewards.filter((r) => !redeemedRewardIds.has(r.id));

  const renderTabContent = () => {
    if (isLoading && !isRefreshing) {
      return (
        <View style={styles.loadingContainer}>
          <ActivityIndicator size="large" color="#3b82f6" />
          <Text style={styles.loadingText}>Loading rewards...</Text>
        </View>
      );
    }

    switch (activeTab) {
      case 'available':
        return (
          <View style={styles.listContainer}>
            {availableRewards.length === 0 ? (
              <View style={styles.emptyContainer}>
                <Text style={styles.emptyEmoji}>🎁</Text>
                <Text style={styles.emptyTitle}>No Available Rewards</Text>
                <Text style={styles.emptyText}>Check back later for new rewards!</Text>
              </View>
            ) : (
              availableRewards.map((reward) => (
                <RewardCard
                  key={reward.id}
                  reward={reward}
                  onRedeem={handleRedeem}
                  showRedeemButton={true}
                />
              ))
            )}
          </View>
        );

      case 'redeemed':
        return (
          <View style={styles.listContainer}>
            {redeemedRewards.length === 0 ? (
              <View style={styles.emptyContainer}>
                <Text style={styles.emptyEmoji}>📦</Text>
                <Text style={styles.emptyTitle}>No Redemptions</Text>
                <Text style={styles.emptyText}>Redeem rewards to see them here!</Text>
              </View>
            ) : (
              redeemedRewards.map((userReward) => (
                <RewardCard
                  key={userReward.id}
                  reward={userReward.reward}
                  userReward={userReward}
                  showRedeemButton={false}
                />
              ))
            )}
          </View>
        );

      case 'create':
        return (
          <View style={styles.createContainer}>
            <TouchableOpacity
              style={styles.createButton}
              onPress={() => setIsCreateModalVisible(true)}
            >
              <Text style={styles.createButtonText}>+ Create New Reward</Text>
            </TouchableOpacity>

            <Text style={styles.hintText}>
              Create custom rewards for yourself or your team!
            </Text>
          </View>
        );
    }
  };

  return (
    <View style={styles.container}>
      <Stack.Screen options={{ title: 'Rewards' }} />

      <View style={styles.tabBar}>
        <TouchableOpacity
          style={[styles.tab, activeTab === 'available' && styles.activeTab]}
          onPress={() => setActiveTab('available')}
        >
          <Text style={[styles.tabText, activeTab === 'available' && styles.activeTabText]}>
            Available
          </Text>
          <View style={styles.badge}>
            <Text style={styles.badgeText}>{availableRewards.length}</Text>
          </View>
        </TouchableOpacity>

        <TouchableOpacity
          style={[styles.tab, activeTab === 'redeemed' && styles.activeTab]}
          onPress={() => setActiveTab('redeemed')}
        >
          <Text style={[styles.tabText, activeTab === 'redeemed' && styles.activeTabText]}>
            Redeemed
          </Text>
          <View style={styles.badge}>
            <Text style={styles.badgeText}>{redeemedRewards.length}</Text>
          </View>
        </TouchableOpacity>

        <TouchableOpacity
          style={[styles.tab, activeTab === 'create' && styles.activeTab]}
          onPress={() => setActiveTab('create')}
        >
          <Text style={[styles.tabText, activeTab === 'create' && styles.activeTabText]}>
            Create
          </Text>
        </TouchableOpacity>
      </View>

      <ScrollView
        style={styles.scrollView}
        refreshControl={
          activeTab !== 'create' ? (
            <RefreshControl
              refreshing={isRefreshing}
              onRefresh={handleRefresh}
              tintColor="#3b82f6"
            />
          ) : undefined
        }
      >
        {renderTabContent()}
      </ScrollView>

      <Modal
        visible={isCreateModalVisible}
        transparent
        animationType="slide"
        onRequestClose={() => setIsCreateModalVisible(false)}
      >
        <View style={styles.modalOverlay}>
          <View style={styles.modalContent}>
            <Text style={styles.modalTitle}>Create Reward</Text>

            <Text style={styles.inputLabel}>Name</Text>
            <TextInput
              style={styles.input}
              value={newReward.name}
              onChangeText={(text) => setNewReward((prev) => ({ ...prev, name: text }))}
              placeholder="Reward name"
            />

            <Text style={styles.inputLabel}>Description</Text>
            <TextInput
              style={[styles.input, styles.textArea]}
              value={newReward.description}
              onChangeText={(text) => setNewReward((prev) => ({ ...prev, description: text }))}
              placeholder="Describe the reward"
              multiline
              numberOfLines={3}
            />

            <Text style={styles.inputLabel}>Type</Text>
            <View style={styles.typeSelector}>
              {(['points', 'badge', 'feature'] as RewardType[]).map((type) => (
                <TouchableOpacity
                  key={type}
                  style={[
                    styles.typeButton,
                    newReward.type === type && styles.typeButtonActive,
                  ]}
                  onPress={() => setNewReward((prev) => ({ ...prev, type }))}
                >
                  <Text
                    style={[
                      styles.typeButtonText,
                      newReward.type === type && styles.typeButtonTextActive,
                    ]}
                  >
                    {type.charAt(0).toUpperCase() + type.slice(1)}
                  </Text>
                </TouchableOpacity>
              ))}
            </View>

            <Text style={styles.inputLabel}>Value (points)</Text>
            <TextInput
              style={styles.input}
              value={newReward.value}
              onChangeText={(text) => setNewReward((prev) => ({ ...prev, value: text }))}
              placeholder="100"
              keyboardType="numeric"
            />

            <View style={styles.modalButtons}>
              <TouchableOpacity
                style={[styles.modalButton, styles.cancelButton]}
                onPress={() => setIsCreateModalVisible(false)}
              >
                <Text style={styles.cancelButtonText}>Cancel</Text>
              </TouchableOpacity>
              <TouchableOpacity
                style={[styles.modalButton, styles.submitButton]}
                onPress={handleCreateReward}
                disabled={createMutation.isPending}
              >
                {createMutation.isPending ? (
                  <ActivityIndicator color="#fff" />
                ) : (
                  <Text style={styles.submitButtonText}>Create</Text>
                )}
              </TouchableOpacity>
            </View>
          </View>
        </View>
      </Modal>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f5f5f5',
  },
  tabBar: {
    flexDirection: 'row',
    backgroundColor: '#fff',
    paddingHorizontal: 8,
    paddingVertical: 8,
    borderBottomWidth: 1,
    borderBottomColor: '#e5e7eb',
  },
  tab: {
    flex: 1,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    paddingVertical: 10,
    borderRadius: 8,
    gap: 6,
  },
  activeTab: {
    backgroundColor: '#dbeafe',
  },
  tabText: {
    fontSize: 14,
    fontWeight: '500',
    color: '#666',
  },
  activeTabText: {
    color: '#3b82f6',
  },
  badge: {
    backgroundColor: '#e5e7eb',
    borderRadius: 10,
    minWidth: 20,
    height: 20,
    alignItems: 'center',
    justifyContent: 'center',
    paddingHorizontal: 6,
  },
  badgeText: {
    fontSize: 12,
    fontWeight: '600',
    color: '#666',
  },
  scrollView: {
    flex: 1,
  },
  listContainer: {
    paddingVertical: 8,
  },
  loadingContainer: {
    paddingVertical: 60,
    alignItems: 'center',
  },
  loadingText: {
    marginTop: 12,
    fontSize: 16,
    color: '#666',
  },
  emptyContainer: {
    paddingVertical: 60,
    alignItems: 'center',
    paddingHorizontal: 40,
  },
  emptyEmoji: {
    fontSize: 48,
    marginBottom: 16,
  },
  emptyTitle: {
    fontSize: 20,
    fontWeight: '600',
    color: '#1a1a1a',
    marginBottom: 8,
  },
  emptyText: {
    fontSize: 14,
    color: '#666',
    textAlign: 'center',
  },
  createContainer: {
    padding: 20,
    alignItems: 'center',
  },
  createButton: {
    backgroundColor: '#3b82f6',
    paddingHorizontal: 24,
    paddingVertical: 14,
    borderRadius: 12,
    marginBottom: 16,
  },
  createButtonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '600',
  },
  hintText: {
    fontSize: 14,
    color: '#666',
    textAlign: 'center',
  },
  modalOverlay: {
    flex: 1,
    backgroundColor: 'rgba(0, 0, 0, 0.5)',
    justifyContent: 'center',
    alignItems: 'center',
    padding: 20,
  },
  modalContent: {
    backgroundColor: '#fff',
    borderRadius: 16,
    padding: 24,
    width: '100%',
    maxWidth: 400,
  },
  modalTitle: {
    fontSize: 20,
    fontWeight: 'bold',
    color: '#1a1a1a',
    marginBottom: 20,
    textAlign: 'center',
  },
  inputLabel: {
    fontSize: 14,
    fontWeight: '500',
    color: '#374151',
    marginBottom: 8,
  },
  input: {
    backgroundColor: '#f3f4f6',
    borderRadius: 8,
    paddingHorizontal: 12,
    paddingVertical: 10,
    fontSize: 16,
    marginBottom: 16,
    borderWidth: 1,
    borderColor: '#e5e7eb',
  },
  textArea: {
    height: 80,
    textAlignVertical: 'top',
  },
  typeSelector: {
    flexDirection: 'row',
    gap: 8,
    marginBottom: 16,
  },
  typeButton: {
    flex: 1,
    paddingVertical: 10,
    borderRadius: 8,
    backgroundColor: '#f3f4f6',
    alignItems: 'center',
  },
  typeButtonActive: {
    backgroundColor: '#3b82f6',
  },
  typeButtonText: {
    fontSize: 13,
    fontWeight: '500',
    color: '#666',
  },
  typeButtonTextActive: {
    color: '#fff',
  },
  modalButtons: {
    flexDirection: 'row',
    gap: 12,
    marginTop: 8,
  },
  modalButton: {
    flex: 1,
    paddingVertical: 12,
    borderRadius: 8,
    alignItems: 'center',
  },
  cancelButton: {
    backgroundColor: '#f3f4f6',
  },
  cancelButtonText: {
    color: '#666',
    fontSize: 16,
    fontWeight: '600',
  },
  submitButton: {
    backgroundColor: '#3b82f6',
  },
  submitButtonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '600',
  },
});
