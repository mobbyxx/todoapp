import { Stack, SplashScreen, useRouter } from 'expo-router';
import { useEffect } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import * as Linking from 'expo-linking';
import { useAuthStore } from '../stores/authStore';
import { useAppSync } from '../hooks/useAppSync';

SplashScreen.preventAutoHideAsync();

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 60000,
      retry: 2,
    },
  },
});

function useDeepLinkHandler() {
  const router = useRouter();
  const { isAuthenticated } = useAuthStore();

  useEffect(() => {
    const handleDeepLink = (url: string) => {
      const { path } = Linking.parse(url);

      if (!path) return;

      if (path.startsWith('invite/') || path.startsWith('/invite/')) {
        const token = path.replace(/^\/?invite\//, '');
        if (token) {
          if (isAuthenticated) {
            router.push(`/invite?token=${token}`);
          } else {
            router.push(`/login?redirect=invite&token=${token}`);
          }
        }
      }
    };

    const subscription = Linking.addEventListener('url', ({ url }) => {
      handleDeepLink(url);
    });

    Linking.getInitialURL().then((url) => {
      if (url) {
        handleDeepLink(url);
      }
    });

    return () => {
      subscription.remove();
    };
  }, [router, isAuthenticated]);
}

function RootLayoutContent() {
  const { isLoading } = useAuthStore();

  useDeepLinkHandler();
  useAppSync();

  useEffect(() => {
    if (!isLoading) {
      SplashScreen.hideAsync();
    }
  }, [isLoading]);

  if (isLoading) {
    return null;
  }

  return (
    <Stack screenOptions={{ headerShown: false }}>
      <Stack.Screen name="login" />
      <Stack.Screen name="register" />
      <Stack.Screen name="invite" />
      <Stack.Screen name="(app)" />
    </Stack>
  );
}

export default function RootLayout() {
  return (
    <QueryClientProvider client={queryClient}>
      <RootLayoutContent />
    </QueryClientProvider>
  );
}
