export type FriendshipStatus = 'PENDING' | 'ACCEPTED' | 'DECLINED' | 'BLOCKED';

export type Friend = {
  avatarUrl?: string | null;
  createdAt: string;
  id: string;
  name: string;
  userId: string;
};

export type FriendRequest = {
  avatarUrl?: string | null;
  createdAt: string;
  id: string;
  name: string;
  userId: string;
};

export type UserSearchResult = {
  avatarUrl?: string | null;
  friendshipStatus: FriendshipStatus | null;
  id: string;
  name: string;
};

export type PublicProfile = {
  avatarUrl?: string | null;
  globalRanking: number | null;
  groupsCount: number;
  joinedAt: string | null;
  name: string;
  predictionsCount: number;
  totalPoints: number;
  userId: string;
};
