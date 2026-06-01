import { Pressable, StyleSheet, Text, View } from 'react-native';

import { EmptyBox } from '../../../../shared/components/EmptyBox';
import { LoadingIndicator } from '../../../../shared/components/LoadingIndicator';
import type { GroupMember, GroupPayment, PaymentStatus } from '../../services/groups';

type GroupAdminMembersProps = {
  isLoadingMembers: boolean;
  loadMembers: () => void;
  members: GroupMember[];
  onRemove: (member: GroupMember) => void;
  onTransferOwnership: (member: GroupMember) => void;
  paymentsByUserID?: Record<string, GroupPayment>;
  removingUserID: string | null;
  transferringOwnerUserID: string | null;
};

const paymentLabels: Record<PaymentStatus, string> = {
  exempt: 'Isento',
  paid: 'Pago',
  pending: 'Pendente',
  refunded: 'Reembolsado',
};

export function GroupAdminMembers({
  isLoadingMembers,
  loadMembers,
  members,
  onRemove,
  onTransferOwnership,
  paymentsByUserID = {},
  removingUserID,
  transferringOwnerUserID,
}: GroupAdminMembersProps) {
  return (
    <View style={styles.card}>
      <View style={styles.header}>
        <View>
          <Text style={styles.cardTitle}>Membros</Text>
          <Text style={styles.cardSubtitle}>Palpiteiros ativos e permissões</Text>
        </View>
        <Pressable onPress={loadMembers} style={styles.refreshButton}>
          <Text style={styles.refreshButtonText}>Atualizar</Text>
        </Pressable>
      </View>

      {isLoadingMembers ? <LoadingIndicator text="Carregando..." /> : null}

      {!isLoadingMembers && members.length === 0 ? (
        <EmptyBox title="Nenhum Palpiteiro." text="Nenhum membro ativo encontrado." />
      ) : null}

      {members.map((member) => {
        const isOwner = member.role === 'owner';
        const payment = paymentsByUserID[member.user_id];

        return (
          <View key={member.user_id} style={styles.memberRow}>
            <View style={styles.memberHeader}>
              <View style={styles.memberInfo}>
                <Text style={styles.memberName}>
                  {member.display_name || `Usuário ${member.user_id.slice(0, 8)}`}
                </Text>
                <Text style={styles.memberMeta}>{isOwner ? 'Dono do grupo' : 'Palpiteiro'}</Text>
              </View>

              <View style={styles.badges}>
                {isOwner ? (
                  <View style={styles.ownerBadge}>
                    <Text style={styles.ownerBadgeText}>Dono</Text>
                  </View>
                ) : null}

                {payment ? (
                  <View style={[styles.paymentBadge, styles[`payment_${payment.status}`]]}>
                    <Text style={[styles.paymentBadgeText, styles[`paymentText_${payment.status}`]]}>
                      {paymentLabels[payment.status]}
                    </Text>
                  </View>
                ) : null}
              </View>
            </View>

            {!isOwner ? (
              <View style={styles.memberActions}>
                <MemberActionButton
                  isLoading={transferringOwnerUserID === member.user_id}
                  loadingLabel="Transferindo..."
                  onPress={() => onTransferOwnership(member)}
                  label="Tornar dono"
                  variant="secondary"
                />
                <MemberActionButton
                  isLoading={removingUserID === member.user_id}
                  loadingLabel="Removendo..."
                  onPress={() => onRemove(member)}
                  label="Remover"
                  variant="danger"
                />
              </View>
            ) : null}
          </View>
        );
      })}
    </View>
  );
}

function MemberActionButton({
  isLoading,
  label,
  loadingLabel,
  onPress,
  variant,
}: {
  isLoading: boolean;
  label: string;
  loadingLabel: string;
  onPress: () => void;
  variant: 'danger' | 'secondary';
}) {
  return (
    <Pressable
      disabled={isLoading}
      onPress={onPress}
      style={[
        styles.memberActionButton,
        variant === 'danger' ? styles.memberActionDanger : styles.memberActionSecondary,
        isLoading && styles.actionDisabled,
      ]}>
      <Text
        style={[
          styles.memberActionText,
          variant === 'danger' ? styles.memberActionDangerText : styles.memberActionSecondaryText,
        ]}>
        {isLoading ? loadingLabel : label}
      </Text>
    </Pressable>
  );
}

const styles = StyleSheet.create({
  card: {
    backgroundColor: '#ffffff',
    borderColor: '#cfe0c9',
    borderRadius: 8,
    borderWidth: 1,
    gap: 16,
    padding: 16,
  },
  header: {
    alignItems: 'center',
    flexDirection: 'row',
    justifyContent: 'space-between',
  },
  cardTitle: {
    color: '#123d2a',
    fontSize: 18,
    fontWeight: '800',
  },
  cardSubtitle: {
    color: '#486654',
    fontSize: 13,
    marginTop: 4,
  },
  refreshButton: {
    backgroundColor: '#f5f8ef',
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
  memberRow: {
    borderTopColor: '#edf3e8',
    borderTopWidth: 1,
    gap: 12,
    paddingTop: 12,
  },
  memberInfo: {
    flex: 1,
  },
  memberHeader: {
    alignItems: 'flex-start',
    flexDirection: 'row',
    gap: 10,
    justifyContent: 'space-between',
  },
  memberActions: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 8,
  },
  memberName: {
    color: '#183f2d',
    fontSize: 14,
    fontWeight: '800',
  },
  memberMeta: {
    color: '#486654',
    fontSize: 12,
    marginTop: 3,
  },
  ownerBadge: {
    backgroundColor: '#edf3e8',
    borderRadius: 8,
    paddingHorizontal: 10,
    paddingVertical: 8,
  },
  ownerBadgeText: {
    color: '#1f7a4a',
    fontSize: 12,
    fontWeight: '800',
  },
  badges: {
    alignItems: 'flex-end',
    gap: 6,
  },
  paymentBadge: {
    borderRadius: 8,
    paddingHorizontal: 10,
    paddingVertical: 8,
  },
  paymentBadgeText: {
    fontSize: 12,
    fontWeight: '800',
  },
  payment_pending: {
    backgroundColor: '#fff7dd',
  },
  payment_paid: {
    backgroundColor: '#e4f6ea',
  },
  payment_exempt: {
    backgroundColor: '#edf3e8',
  },
  payment_refunded: {
    backgroundColor: '#fde9e7',
  },
  paymentText_pending: {
    color: '#8a5d00',
  },
  paymentText_paid: {
    color: '#1f7a4a',
  },
  paymentText_exempt: {
    color: '#486654',
  },
  paymentText_refunded: {
    color: '#b23b32',
  },
  memberActionButton: {
    alignItems: 'center',
    borderRadius: 8,
    borderWidth: 1,
    flexGrow: 1,
    justifyContent: 'center',
    minHeight: 42,
    paddingHorizontal: 12,
    paddingVertical: 9,
  },
  memberActionSecondary: {
    backgroundColor: '#f5f8ef',
    borderColor: '#cfe0c9',
  },
  memberActionDanger: {
    backgroundColor: '#fff6f4',
    borderColor: '#efc3bd',
  },
  memberActionText: {
    fontSize: 12,
    fontWeight: '800',
    textAlign: 'center',
  },
  memberActionSecondaryText: {
    color: '#1f7a4a',
  },
  memberActionDangerText: {
    color: '#a03222',
  },
  actionDisabled: {
    opacity: 0.65,
  },
});
