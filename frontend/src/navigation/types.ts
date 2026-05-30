import type { Group } from '../features/groups/services/groups';

export type AuthScreenName = 'login' | 'signup';

export type AppScreenName =
  | 'home'
  | 'create-group'
  | 'group-detail'
  | 'group-admin'
  | 'group-members'
  | 'group-member-detail'
  | 'profile';

export type RootNavigationState = {
  appScreen: AppScreenName;
  authScreen: AuthScreenName;
  selectedGroup: Group | null;
};
