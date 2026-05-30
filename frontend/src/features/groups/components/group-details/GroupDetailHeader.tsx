import { Pressable, StyleSheet, Text, View } from 'react-native';

import { NotificationBanner } from '../../../../shared/components/NotificationBanner';
import type { Group } from '../../services/groups';
import type { GroupDetailTab } from '../../types';
import { BackButton } from '../../../../shared/components/BackButton';

type Props = {
  activeTab: GroupDetailTab;
  group: Group;
  isLeavingGroup: boolean;
  notificationMessage: string | null;
  onBack: () => void;
  onChangeTab: (tab: GroupDetailTab) => void;
  onLeaveGroup: () => void;
  onOpenAdmin: () => void;
};

export function GroupDetailHeader({
  activeTab,
  group,
  isLeavingGroup,
  notificationMessage,
  onBack,
  onChangeTab,
  onLeaveGroup,
  onOpenAdmin,
}: Props) {
  return (
    <View style={styles.headerBlock}>
      <View style={styles.topBar}>
        <BackButton onPress={onBack} />

        {group.role === 'owner' ? (
          <Pressable onPress={onOpenAdmin} style={styles.adminButton}>
            <Text style={styles.adminButtonText}>Admin</Text>
          </Pressable>
        ) : (
          <Pressable
            disabled={isLeavingGroup}
            onPress={onLeaveGroup}
            style={[styles.leaveButton, isLeavingGroup && styles.buttonDisabled]}>
            <Text style={styles.leaveButtonText}>
              {isLeavingGroup ? 'Saindo...' : 'Sair'}
            </Text>
          </Pressable>
        )}
      </View>

      <View style={styles.titleBlock}>
        <Text style={styles.title}>{group.name}</Text>
        <Text style={styles.subtitle}>
          {group.match_scope === 'all'
            ? 'Todos os jogos da Copa'
            : `Seleções: ${group.selected_teams.join(', ')}`}
        </Text>
      </View>

      <NotificationBanner message={notificationMessage} />

      <View style={styles.tabs}>
        <Pressable
          onPress={() => onChangeTab('matches')}
          style={[styles.tabButton, activeTab === 'matches' && styles.tabButtonActive]}>
          <Text
            style={[styles.tabButtonText, activeTab === 'matches' && styles.tabButtonTextActive]}>
            Jogos e palpites
          </Text>
        </Pressable>

        <Pressable
          onPress={() => onChangeTab('ranking')}
          style={[styles.tabButton, activeTab === 'ranking' && styles.tabButtonActive]}>
          <Text
            style={[styles.tabButtonText, activeTab === 'ranking' && styles.tabButtonTextActive]}>
            Ranking
          </Text>
        </Pressable>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  headerBlock: {
    paddingTop: 12,
  },
  topBar: {
    alignItems: 'center',
    flexDirection: 'row',
    justifyContent: 'space-between',
  },
  backButton: {
    alignItems: 'center',
    backgroundColor: '#ffffff',
    borderColor: '#d9e7d4',
    borderRadius: 22,
    borderWidth: 1,
    height: 44,
    justifyContent: 'center',
    width: 44,
  },
  backButtonText: {
    color: '#1f7a4a',
    fontSize: 34,
    fontWeight: '600',
    lineHeight: 38,
  },
  adminButton: {
    backgroundColor: '#ffffff',
    borderColor: '#1f7a4a',
    borderRadius: 8,
    borderWidth: 1,
    paddingHorizontal: 14,
    paddingVertical: 10,
  },
  adminButtonText: {
    color: '#1f7a4a',
    fontSize: 14,
    fontWeight: '800',
  },
  leaveButton: {
    backgroundColor: '#ffffff',
    borderColor: '#a03222',
    borderRadius: 8,
    borderWidth: 1,
    paddingHorizontal: 14,
    paddingVertical: 10,
  },
  leaveButtonText: {
    color: '#a03222',
    fontSize: 14,
    fontWeight: '800',
  },
  buttonDisabled: {
    opacity: 0.7,
  },
  titleBlock: {
    paddingTop: 12,
  },
  title: {
    color: '#123d2a',
    fontSize: 34,
    fontWeight: '800',
  },
  subtitle: {
    color: '#486654',
    fontSize: 15,
    lineHeight: 22,
    marginTop: 8,
  },
  tabs: {
    backgroundColor: '#edf3e8',
    borderRadius: 8,
    flexDirection: 'row',
    gap: 6,
    marginTop: 20,
    padding: 6,
  },
  tabButton: {
    alignItems: 'center',
    borderRadius: 7,
    flex: 1,
    justifyContent: 'center',
    minHeight: 44,
  },
  tabButtonActive: {
    backgroundColor: '#ffffff',
    shadowColor: '#1e5c39',
    shadowOffset: { height: 4, width: 0 },
    shadowOpacity: 0.08,
    shadowRadius: 8,
  },
  tabButtonText: {
    color: '#486654',
    fontSize: 13,
    fontWeight: '800',
  },
  tabButtonTextActive: {
    color: '#1f7a4a',
  },
});
