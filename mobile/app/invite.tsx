import { useEffect, useState } from 'react';
import {
  View,
  Text,
  TouchableOpacity,
  StyleSheet,
  ActivityIndicator,
  ScrollView,
  SafeAreaView,
} from 'react-native';
import { router, useLocalSearchParams } from 'expo-router';
import { useAuthStore } from '../stores/authStore';
import {
  Invitation,
  getInvitationByToken,
  acceptInvitation,
  rejectInvitation,
} from '../services/invitations';

type InvitationState =
  | { status: 'loading' }
  | { status: 'loaded'; invitation: Invitation }
  | { status: 'error'; message: string }
  | { status: 'accepted'; invitation: Invitation }
  | { status: 'rejected'; invitation: Invitation };

export default function InviteScreen() {
  const { token } = useLocalSearchParams<{ token?: string }>();
  const { isAuthenticated } = useAuthStore();
  const [state, setState] = useState<InvitationState>({ status: 'loading' });

  useEffect(() => {
    if (!token) {
      setState({ status: 'error', message: 'Invalid invitation link' });
      return;
    }

    if (!isAuthenticated) {
      router.replace(`/login?redirect=invite&token=${token}`);
      return;
    }

    loadInvitation();
  }, [token, isAuthenticated]);

  const loadInvitation = async () => {
    if (!token) return;

    try {
      const invitation = await getInvitationByToken(token);

      if (invitation.status !== 'pending') {
        setState({
          status: 'error',
          message: `This invitation has already been ${invitation.status}`,
        });
        return;
      }

      setState({ status: 'loaded', invitation });
    } catch (error) {
      setState({
        status: 'error',
        message: error instanceof Error ? error.message : 'Failed to load invitation',
      });
    }
  };

  const handleAccept = async () => {
    if (!token || state.status !== 'loaded') return;

    setState({ status: 'loading' });

    try {
      const result = await acceptInvitation(token);
      setState({ status: 'accepted', invitation: result.invitation });
    } catch (error) {
      setState({
        status: 'error',
        message: error instanceof Error ? error.message : 'Failed to accept invitation',
      });
    }
  };

  const handleReject = async () => {
    if (!token || state.status !== 'loaded') return;

    setState({ status: 'loading' });

    try {
      const result = await rejectInvitation(token);
      setState({ status: 'rejected', invitation: result.invitation });
    } catch (error) {
      setState({
        status: 'error',
        message: error instanceof Error ? error.message : 'Failed to reject invitation',
      });
    }
  };

  const handleGoHome = () => {
    router.replace('/(app)');
  };

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    });
  };

  const renderContent = () => {
    switch (state.status) {
      case 'loading':
        return (
          <View style={styles.centered}>
            <ActivityIndicator size="large" color="#007AFF" />
            <Text style={styles.loadingText}>Loading invitation...</Text>
          </View>
        );

      case 'error':
        return (
          <View style={styles.centered}>
            <View style={styles.iconContainer}>
              <Text style={styles.errorIcon}>⚠️</Text>
            </View>
            <Text style={styles.errorTitle}>Oops!</Text>
            <Text style={styles.errorMessage}>{state.message}</Text>
            <TouchableOpacity style={styles.primaryButton} onPress={handleGoHome}>
              <Text style={styles.primaryButtonText}>Go Home</Text>
            </TouchableOpacity>
          </View>
        );

      case 'accepted':
        return (
          <View style={styles.centered}>
            <View style={styles.iconContainer}>
              <Text style={styles.successIcon}>✓</Text>
            </View>
            <Text style={styles.successTitle}>Invitation Accepted!</Text>
            <Text style={styles.successMessage}>
              You are now connected with {state.invitation.sender_name}
            </Text>
            <TouchableOpacity style={styles.primaryButton} onPress={handleGoHome}>
              <Text style={styles.primaryButtonText}>Go to Dashboard</Text>
            </TouchableOpacity>
          </View>
        );

      case 'rejected':
        return (
          <View style={styles.centered}>
            <View style={styles.iconContainer}>
              <Text style={styles.rejectIcon}>✕</Text>
            </View>
            <Text style={styles.rejectTitle}>Invitation Declined</Text>
            <Text style={styles.rejectMessage}>
              You have declined the invitation from {state.invitation.sender_name}
            </Text>
            <TouchableOpacity style={styles.primaryButton} onPress={handleGoHome}>
              <Text style={styles.primaryButtonText}>Go Home</Text>
            </TouchableOpacity>
          </View>
        );

      case 'loaded':
        return (
          <View style={styles.content}>
            <View style={styles.header}>
              <View style={styles.avatarContainer}>
                <Text style={styles.avatarText}>
                  {state.invitation.sender_name.charAt(0).toUpperCase()}
                </Text>
              </View>
              <Text style={styles.title}>Connection Invitation</Text>
              <Text style={styles.subtitle}>
                {state.invitation.sender_name} wants to connect with you
              </Text>
            </View>

            <View style={styles.detailsCard}>
              <View style={styles.detailRow}>
                <Text style={styles.detailLabel}>From</Text>
                <Text style={styles.detailValue}>{state.invitation.sender_email}</Text>
              </View>
              <View style={styles.detailRow}>
                <Text style={styles.detailLabel}>Sent</Text>
                <Text style={styles.detailValue}>
                  {formatDate(state.invitation.created_at)}
                </Text>
              </View>
              <View style={styles.detailRow}>
                <Text style={styles.detailLabel}>Expires</Text>
                <Text style={styles.detailValue}>
                  {formatDate(state.invitation.expires_at)}
                </Text>
              </View>
            </View>

            <View style={styles.actions}>
              <TouchableOpacity
                style={[styles.button, styles.acceptButton]}
                onPress={handleAccept}
              >
                <Text style={styles.acceptButtonText}>Accept</Text>
              </TouchableOpacity>
              <TouchableOpacity
                style={[styles.button, styles.rejectButton]}
                onPress={handleReject}
              >
                <Text style={styles.rejectButtonText}>Decline</Text>
              </TouchableOpacity>
            </View>
          </View>
        );
    }
  };

  return (
    <SafeAreaView style={styles.container}>
      <ScrollView
        contentContainerStyle={styles.scrollContent}
        keyboardShouldPersistTaps="handled"
      >
        {renderContent()}
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#fff',
  },
  scrollContent: {
    flexGrow: 1,
  },
  centered: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 24,
  },
  content: {
    flex: 1,
    padding: 24,
  },
  header: {
    alignItems: 'center',
    marginBottom: 32,
    marginTop: 48,
  },
  avatarContainer: {
    width: 80,
    height: 80,
    borderRadius: 40,
    backgroundColor: '#007AFF',
    justifyContent: 'center',
    alignItems: 'center',
    marginBottom: 16,
  },
  avatarText: {
    fontSize: 32,
    fontWeight: 'bold',
    color: '#fff',
  },
  title: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#1a1a1a',
    marginBottom: 8,
  },
  subtitle: {
    fontSize: 16,
    color: '#666',
    textAlign: 'center',
  },
  detailsCard: {
    backgroundColor: '#f9fafb',
    borderRadius: 16,
    padding: 20,
    marginBottom: 32,
  },
  detailRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    paddingVertical: 12,
    borderBottomWidth: 1,
    borderBottomColor: '#e5e7eb',
  },
  detailLabel: {
    fontSize: 14,
    color: '#6b7280',
    fontWeight: '500',
  },
  detailValue: {
    fontSize: 14,
    color: '#1a1a1a',
    fontWeight: '500',
  },
  actions: {
    gap: 12,
  },
  button: {
    borderRadius: 12,
    paddingVertical: 16,
    alignItems: 'center',
  },
  acceptButton: {
    backgroundColor: '#22c55e',
  },
  acceptButtonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '600',
  },
  rejectButton: {
    backgroundColor: '#f3f4f6',
    borderWidth: 1,
    borderColor: '#e5e7eb',
  },
  rejectButtonText: {
    color: '#6b7280',
    fontSize: 16,
    fontWeight: '600',
  },
  loadingText: {
    marginTop: 16,
    fontSize: 16,
    color: '#666',
  },
  iconContainer: {
    width: 80,
    height: 80,
    borderRadius: 40,
    justifyContent: 'center',
    alignItems: 'center',
    marginBottom: 24,
  },
  errorIcon: {
    fontSize: 48,
  },
  errorTitle: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#dc2626',
    marginBottom: 8,
  },
  errorMessage: {
    fontSize: 16,
    color: '#666',
    textAlign: 'center',
    marginBottom: 24,
  },
  successIcon: {
    fontSize: 40,
    color: '#22c55e',
    fontWeight: 'bold',
  },
  successTitle: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#16a34a',
    marginBottom: 8,
  },
  successMessage: {
    fontSize: 16,
    color: '#666',
    textAlign: 'center',
    marginBottom: 24,
  },
  rejectIcon: {
    fontSize: 40,
    color: '#6b7280',
    fontWeight: 'bold',
  },
  rejectTitle: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#374151',
    marginBottom: 8,
  },
  rejectMessage: {
    fontSize: 16,
    color: '#666',
    textAlign: 'center',
    marginBottom: 24,
  },
  primaryButton: {
    backgroundColor: '#007AFF',
    borderRadius: 12,
    paddingVertical: 16,
    paddingHorizontal: 32,
    alignItems: 'center',
  },
  primaryButtonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '600',
  },
});
