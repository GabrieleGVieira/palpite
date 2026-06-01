import * as Clipboard from 'expo-clipboard';
import { useState } from 'react';
import { Pressable, StyleSheet, Text, View } from 'react-native';
import type { Group } from '../../services/groups';
import { LoadingIndicator } from '../../../../shared/components/LoadingIndicator';

type GroupListSectionProps = {
  groups: Group[];
  isLoadingGroups: boolean;
  groupsError: string | null;
  onRefresh: () => void;
  onOpenGroup: (group: Group) => void;
};

export function GroupListSection({
  groups,
  isLoadingGroups,
  groupsError,
  onRefresh,
  onOpenGroup,
}: GroupListSectionProps) {
  const [copiedGroupID, setCopiedGroupID] = useState<string | null>(null);

  async function copyInviteCode(group: Group) {
    await Clipboard.setStringAsync(group.invite_code);
    setCopiedGroupID(group.id);
    setTimeout(() => {
      setCopiedGroupID((currentGroupID) => (currentGroupID === group.id ? null : currentGroupID));
    }, 1800);
  }

  return (
    <View style={styles.section}>
      <View style={styles.sectionHeader}>
        <View>
          <Text style={styles.sectionTitle}>Meus grupos</Text>
          <Text style={styles.sectionSubtitle}>Bolões em que você participa</Text>
        </View>
        <Pressable onPress={onRefresh} style={styles.refreshButton}>
          <Text style={styles.refreshButtonText}>Atualizar</Text>
        </Pressable>
      </View>

      {isLoadingGroups ? <LoadingIndicator text="Carregando grupos..." /> : null}

      {groupsError ? <Text style={styles.errorText}>{groupsError}</Text> : null}

      {!isLoadingGroups && !groupsError && groups.length === 0 ? (
        <View style={styles.emptyBox}>
          <Text style={styles.emptyTitle}>Nenhum grupo ainda</Text>
          <Text style={styles.emptyText}>
            Crie seu primeiro bolão da Copa para convidar sua turma.
          </Text>
        </View>
      ) : null}

      {groups.map((group) => {
        const isPendingMembership = group.status === 'pending';

        return (
        <Pressable
          key={group.id}
          disabled={isPendingMembership}
          onPress={() => onOpenGroup(group)}
          style={[styles.groupCard, isPendingMembership && styles.groupCardPending]}>
          <View style={styles.groupCardHeader}>
            <View style={styles.groupTitleBlock}>
              <Text style={styles.groupName}>{group.name}</Text>
              <Text style={styles.groupMeta}>
                {isPendingMembership
                  ? 'Aguardando aprovação do dono'
                  : `${group.role === 'owner' ? 'Dono' : 'Membro'} · ${group.member_count} Palpiteiro${group.member_count === 1 ? '' : 's'}`}
              </Text>
            </View>
            {isPendingMembership ? (
              <View style={styles.pendingBadge}>
                <Text style={styles.pendingBadgeText}>Pendente</Text>
              </View>
            ) : (
              <Pressable
                onLongPress={() => {
                  void copyInviteCode(group);
                }}
                style={styles.inviteBadge}>
                <Text style={styles.inviteBadgeLabel}>Convite</Text>
                <Text style={styles.inviteBadgeCode}>
                  {copiedGroupID === group.id ? 'Copiado' : group.invite_code}
                </Text>
              </Pressable>
            )}
          </View>

          {group.description ? (
            <Text style={styles.groupDescription}>{group.description}</Text>
          ) : null}

          <Text style={styles.groupScope}>
            {group.match_scope === 'all'
              ? 'Todos os jogos da Copa'
              : `Seleções: ${group.selected_teams.join(', ')}`}
          </Text>

          {group.role === 'owner' && group.pending_requests_count > 0 ? (
            <View style={styles.requestsSummaryBox}>
              <Text style={styles.requestsSummaryText}>
                {group.pending_requests_count}{' '}
                {group.pending_requests_count === 1
                  ? 'solicitação pendente'
                  : 'solicitações pendentes'}
              </Text>
              <Text style={styles.requestsSummaryHint}>Abra o admin para analisar.</Text>
            </View>
          ) : null}
        </Pressable>
      );
      })}
    </View>
  );
}

const styles = StyleSheet.create({
  section: {
    gap: 12,
  },
  sectionHeader: {
    alignItems: 'center',
    flexDirection: 'row',
    justifyContent: 'space-between',
  },
  sectionTitle: {
    color: '#123d2a',
    fontSize: 22,
    fontWeight: '800',
  },
  sectionSubtitle: {
    color: '#486654',
    fontSize: 13,
    marginTop: 3,
  },
  refreshButton: {
    backgroundColor: '#ffffff',
    borderColor: '#cfe0c9',
    borderRadius: 8,
    borderWidth: 1,
    paddingHorizontal: 12,
    paddingVertical: 8,
  },
  refreshButtonText: {
    color: '#1f7a4a',
    fontSize: 13,
    fontWeight: '800',
  },
  loadingBox: {
    alignItems: 'center',
    backgroundColor: '#ffffff',
    borderColor: '#cfe0c9',
    borderRadius: 8,
    borderWidth: 1,
    gap: 8,
    padding: 18,
  },
  loadingText: {
    color: '#486654',
    fontSize: 14,
  },
  errorText: {
    color: '#a03222',
    fontSize: 13,
    lineHeight: 18,
  },
  emptyBox: {
    backgroundColor: '#ffffff',
    borderColor: '#cfe0c9',
    borderRadius: 8,
    borderWidth: 1,
    padding: 18,
  },
  emptyTitle: {
    color: '#123d2a',
    fontSize: 17,
    fontWeight: '800',
  },
  emptyText: {
    color: '#486654',
    fontSize: 14,
    lineHeight: 20,
    marginTop: 6,
  },
  groupCard: {
    backgroundColor: '#ffffff',
    borderColor: '#cfe0c9',
    borderRadius: 8,
    borderWidth: 1,
    padding: 16,
  },
  groupCardPending: {
    borderColor: '#e5c76a',
    opacity: 0.92,
  },
  groupCardHeader: {
    alignItems: 'flex-start',
    flexDirection: 'row',
    gap: 12,
    justifyContent: 'space-between',
  },
  groupTitleBlock: {
    flex: 1,
  },
  groupName: {
    color: '#123d2a',
    fontSize: 18,
    fontWeight: '800',
  },
  groupMeta: {
    color: '#486654',
    fontSize: 13,
    marginTop: 4,
  },
  inviteBadge: {
    alignItems: 'center',
    backgroundColor: '#edf3e8',
    borderRadius: 8,
    paddingHorizontal: 10,
    paddingVertical: 8,
  },
  inviteBadgeLabel: {
    color: '#486654',
    fontSize: 10,
    fontWeight: '800',
    textTransform: 'uppercase',
  },
  inviteBadgeCode: {
    color: '#1f7a4a',
    fontSize: 14,
    fontWeight: '800',
    marginTop: 2,
  },
  pendingBadge: {
    alignItems: 'center',
    backgroundColor: '#fff7dd',
    borderRadius: 8,
    paddingHorizontal: 10,
    paddingVertical: 8,
  },
  pendingBadgeText: {
    color: '#8a5d00',
    fontSize: 12,
    fontWeight: '800',
  },
  groupDescription: {
    color: '#486654',
    fontSize: 14,
    lineHeight: 20,
    marginTop: 12,
  },
  groupScope: {
    color: '#1f7a4a',
    fontSize: 13,
    fontWeight: '800',
    marginTop: 12,
  },
  requestsSummaryBox: {
    backgroundColor: '#edf3e8',
    borderRadius: 8,
    marginTop: 14,
    padding: 12,
  },
  requestsSummaryText: {
    color: '#123d2a',
    fontSize: 14,
    fontWeight: '800',
  },
  requestsSummaryHint: {
    color: '#486654',
    fontSize: 12,
    marginTop: 3,
  },
});
