import { useState } from 'react';
import { ActivityIndicator, StyleSheet, View } from 'react-native';

import { useAuth } from './src/hooks/useAuth';
import { CreateGroupScreen } from './src/screens/CreateGroupScreen';
import { HomeScreen } from './src/screens/HomeScreen';
import { LoginScreen } from './src/screens/LoginScreen';
import { OnboardingScreen } from './src/screens/OnboardingScreen';
import { SignupScreen } from './src/screens/SignupScreen';
import { AuthProvider } from './src/store/AuthProvider';

function AppContent() {
  const { isLoading, session } = useAuth();
  const [hasSeenOnboarding, setHasSeenOnboarding] = useState(false);
  const [authScreen, setAuthScreen] = useState<'login' | 'signup'>('login');
  const [appScreen, setAppScreen] = useState<'home' | 'create-group'>('home');

  if (isLoading) {
    return (
      <View style={styles.loadingContainer}>
        <ActivityIndicator color="#1f7a4a" />
      </View>
    );
  }

  if (session) {
    if (appScreen === 'create-group') {
      return <CreateGroupScreen onBack={() => setAppScreen('home')} />;
    }

    return <HomeScreen onCreateGroup={() => setAppScreen('create-group')} />;
  }

  if (!hasSeenOnboarding) {
    return <OnboardingScreen onFinish={() => setHasSeenOnboarding(true)} />;
  }

  if (authScreen === 'signup') {
    return <SignupScreen onBackToLogin={() => setAuthScreen('login')} />;
  }

  return <LoginScreen onCreateAccount={() => setAuthScreen('signup')} />;
}

export default function App() {
  return (
    <AuthProvider>
      <AppContent />
    </AuthProvider>
  );
}

const styles = StyleSheet.create({
  loadingContainer: {
    alignItems: 'center',
    backgroundColor: '#f5f8ef',
    flex: 1,
    justifyContent: 'center',
  },
});
