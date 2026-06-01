import { useState } from 'react';
import { ActivityIndicator, StyleSheet, View } from 'react-native';

import { useAuth } from '../features/auth/hooks/useAuth';
import { LoginScreen } from '../features/auth/screens/LoginScreen';
import { SignupScreen } from '../features/auth/screens/SignupScreen';
import { FriendsScreen } from '../features/friends/screens/FriendsScreen';
import { PublicProfileScreen } from '../features/friends/screens/PublicProfileScreen';
import { ProfileScreen } from '../features/account/screens/ProfileScreen';
import { CreateGroupScreen } from '../features/groups/screens/CreateGroupScreen';
import { GroupAdminScreen } from '../features/groups/screens/GroupAdminScreen';
import { GroupDetailScreen } from '../features/groups/screens/GroupDetailScreen';
import { GroupMemberDetailScreen } from '../features/groups/screens/GroupMemberDetailScreen';
import { GroupMembersScreen } from '../features/groups/screens/GroupMembersScreen';
import { HomeScreen } from '../features/groups/screens/HomeScreen';
import type { Group, GroupMember } from '../features/groups/services/groups';
import { OnboardingScreen } from '../features/onboarding/screens/OnboardingScreen';
import { colors } from '../shared/theme';
import type { AppScreenName, AuthScreenName } from './types';

export function AppNavigator() {
  const { isLoading, session } = useAuth();
  const [hasSeenOnboarding, setHasSeenOnboarding] = useState(false);
  const [authScreen, setAuthScreen] = useState<AuthScreenName>('login');
  const [appScreen, setAppScreen] = useState<AppScreenName>('home');
  const [selectedGroup, setSelectedGroup] = useState<Group | null>(null);
  const [selectedMember, setSelectedMember] = useState<GroupMember | null>(null);
  const [selectedPublicProfileUserID, setSelectedPublicProfileUserID] = useState<string | null>(null);

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
    fallbackName: session.user.user_metadata.full_name as string | undefined,
    selectedMember,
    selectedPublicProfileUserID,
    selectedGroup,
    setAppScreen,
    setSelectedMember,
    setSelectedPublicProfileUserID,
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
  fallbackName?: string;
  selectedMember: GroupMember | null;
  selectedPublicProfileUserID: string | null;
  selectedGroup: Group | null;
  setAppScreen: (screen: AppScreenName) => void;
  setSelectedMember: (member: GroupMember | null) => void;
  setSelectedPublicProfileUserID: (userID: string | null) => void;
  setSelectedGroup: (group: Group | null) => void;
};

function renderAppFlow({
  appScreen,
  fallbackName,
  selectedMember,
  selectedPublicProfileUserID,
  selectedGroup,
  setAppScreen,
  setSelectedMember,
  setSelectedPublicProfileUserID,
  setSelectedGroup,
}: AppFlowParams) {
  if (appScreen === 'profile') {
    return <ProfileScreen fallbackName={fallbackName} onBack={() => setAppScreen('home')} />;
  }

  if (appScreen === 'friends') {
    return (
      <FriendsScreen
        onBack={() => setAppScreen('home')}
        onOpenProfile={(userID) => {
          setSelectedPublicProfileUserID(userID);
          setAppScreen('public-profile');
        }}
      />
    );
  }

  if (appScreen === 'public-profile' && selectedPublicProfileUserID) {
    return (
      <PublicProfileScreen
        onBack={() => setAppScreen('friends')}
        userID={selectedPublicProfileUserID}
      />
    );
  }

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
        onGroupLeft={() => {
          setSelectedGroup(null);
          setAppScreen('home');
        }}
        onOpenAdmin={() => setAppScreen('group-admin')}
        onOpenMembers={() => setAppScreen('group-members')}
      />
    );
  }

  if (appScreen === 'group-members' && selectedGroup) {
    return (
      <GroupMembersScreen
        group={selectedGroup}
        onBack={() => setAppScreen('group-detail')}
        onOpenMember={(member) => {
          setSelectedMember(member);
          setAppScreen('group-member-detail');
        }}
      />
    );
  }

  if (appScreen === 'group-member-detail' && selectedGroup && selectedMember) {
    return (
      <GroupMemberDetailScreen
        group={selectedGroup}
        member={selectedMember}
        onBack={() => setAppScreen('group-members')}
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
      onOpenFriends={() => setAppScreen('friends')}
      onOpenGroup={(group) => {
        setSelectedGroup(group);
        setAppScreen('group-detail');
      }}
      onOpenProfile={() => setAppScreen('profile')}
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
