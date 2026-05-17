import { useState } from 'react';
import { ActivityIndicator, StyleSheet, View } from 'react-native';

import { useAuth } from './src/hooks/useAuth';
import { CreateGroupScreen } from './src/screens/CreateGroupScreen';
import { GroupAdminScreen } from './src/screens/GroupAdminScreen';
import { GroupDetailScreen } from './src/screens/GroupDetailScreen';
import { HomeScreen } from './src/screens/HomeScreen';
import { LoginScreen } from './src/screens/LoginScreen';
import { OnboardingScreen } from './src/screens/OnboardingScreen';
import { SignupScreen } from './src/screens/SignupScreen';
import type { Group } from './src/services/groups';
import { AuthProvider } from './src/store/AuthProvider';

function AppContent() {
  const { isLoading, session } = useAuth();
  const [hasSeenOnboarding, setHasSeenOnboarding] = useState(false);
  const [authScreen, setAuthScreen] = useState<'login' | 'signup'>('login');
  const [appScreen, setAppScreen] = useState<
    'home' | 'create-group' | 'group-detail' | 'group-admin'
  >('home');
  const [selectedGroup, setSelectedGroup] = useState<Group | null>(null);

  if (isLoading) {
    return (
      <View style={styles.loadingContainer}>
        <ActivityIndicator color="#1f7a4a" />
      </View>
    );
  }

  if (session) {
    if (appScreen === 'create-group') {
      return (
        <CreateGroupScreen
          onBack={() => setAppScreen('home')}
          onGroupCreated={() => setAppScreen('home')}
        />
      );
    }

    if (appScreen === 'group-detail' && selectedGroup) {
      return (
        <GroupDetailScreen
          group={selectedGroup}
          onBack={() => setAppScreen('home')}
          onOpenAdmin={() => setAppScreen('group-admin')}
        />
      );
    }

    if (appScreen === 'group-admin' && selectedGroup) {
      return (
        <GroupAdminScreen
          group={selectedGroup}
          onBack={() => setAppScreen('group-detail')}
          onGroupUpdated={setSelectedGroup}
        />
      );
    }

    return (
      <HomeScreen
        onCreateGroup={() => setAppScreen('create-group')}
        onOpenGroup={(group) => {
          setSelectedGroup(group);
          setAppScreen('group-detail');
        }}
      />
    );
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
