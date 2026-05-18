import { useState } from 'react';
import { ActivityIndicator, StyleSheet, View } from 'react-native';

import { useAuth } from '../features/auth/hooks/useAuth';
import { LoginScreen } from '../features/auth/screens/LoginScreen';
import { SignupScreen } from '../features/auth/screens/SignupScreen';
import { CreateGroupScreen } from '../features/groups/screens/CreateGroupScreen';
import { GroupAdminScreen } from '../features/groups/screens/GroupAdminScreen';
import { GroupDetailScreen } from '../features/groups/screens/GroupDetailScreen';
import { HomeScreen } from '../features/groups/screens/HomeScreen';
import type { Group } from '../features/groups/services/groups';
import { OnboardingScreen } from '../features/onboarding/screens/OnboardingScreen';
import { colors } from '../shared/theme';
import type { AppScreenName, AuthScreenName } from './types';

export function AppNavigator() {
  const { isLoading, session } = useAuth();
  const [hasSeenOnboarding, setHasSeenOnboarding] = useState(false);
  const [authScreen, setAuthScreen] = useState<AuthScreenName>('login');
  const [appScreen, setAppScreen] = useState<AppScreenName>('home');
  const [selectedGroup, setSelectedGroup] = useState<Group | null>(null);

  if (isLoading) {
    return (
      <View style={styles.loadingContainer}>
        <ActivityIndicator color={colors.primary} />
      </View>
    );
  }

  if (!session) {
    return renderAuthFlow({
      authScreen,
      hasSeenOnboarding,
      setAuthScreen,
      setHasSeenOnboarding,
    });
  }

  return renderAppFlow({
    appScreen,
    selectedGroup,
    setAppScreen,
    setSelectedGroup,
  });
}

type AuthFlowParams = {
  authScreen: AuthScreenName;
  hasSeenOnboarding: boolean;
  setAuthScreen: (screen: AuthScreenName) => void;
  setHasSeenOnboarding: (hasSeen: boolean) => void;
};

function renderAuthFlow({
  authScreen,
  hasSeenOnboarding,
  setAuthScreen,
  setHasSeenOnboarding,
}: AuthFlowParams) {
  if (!hasSeenOnboarding) {
    return <OnboardingScreen onFinish={() => setHasSeenOnboarding(true)} />;
  }

  if (authScreen === 'signup') {
    return <SignupScreen onBackToLogin={() => setAuthScreen('login')} />;
  }

  return <LoginScreen onCreateAccount={() => setAuthScreen('signup')} />;
}

type AppFlowParams = {
  appScreen: AppScreenName;
  selectedGroup: Group | null;
  setAppScreen: (screen: AppScreenName) => void;
  setSelectedGroup: (group: Group | null) => void;
};

function renderAppFlow({
  appScreen,
  selectedGroup,
  setAppScreen,
  setSelectedGroup,
}: AppFlowParams) {
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

const styles = StyleSheet.create({
  loadingContainer: {
    alignItems: 'center',
    backgroundColor: colors.background,
    flex: 1,
    justifyContent: 'center',
  },
});
