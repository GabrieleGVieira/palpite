import { Pressable, StyleSheet, Text, View } from 'react-native';

import { EmptyBox } from '../global/EmptyBox';
import { GroupAdminRequestItem } from './GroupAdminRequestItem';
import { LoadingIndicator } from '../global/LoadingIndicator';
import type { JoinRequest } from '../../services/groups';

type GroupAdminRequestsProps = {
  approvingUserID: string | null;
  isLoadingRequests: boolean;
  loadRequests: () => void;
  onApprove: (request: JoinRequest) => void;
  requests: JoinRequest[];
};

export function GroupAdminRequests({
  approvingUserID,
  isLoadingRequests,
  loadRequests,
  onApprove,
  requests,
}: GroupAdminRequestsProps) {
  return (
    <View style={styles.card}>
      <View style={styles.requestsHeader}>
        <View>
          <Text style={styles.cardTitle}>Solicitações</Text>
          <Text style={styles.cardSubtitle}>Usuários aguardando aceite</Text>
        </View>
        <Pressable onPress={loadRequests} style={styles.refreshButton}>
          <Text style={styles.refreshButtonText}>Atualizar</Text>
        </Pressable>
      </View>

      {isLoadingRequests ? <LoadingIndicator text="Carregando..." /> : null}

      {!isLoadingRequests && requests.length === 0 ? (
        <EmptyBox
          title="Nenhuma solicitação pendente."
          text="Nenhum pedido de entrada encontrado."
        />
      ) : null}

      {requests.map((request) => (
        <GroupAdminRequestItem
          key={request.user_id}
          isApproving={approvingUserID === request.user_id}
          onApprove={onApprove}
          request={request}
        />
      ))}
    </View>
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
  requestsHeader: {
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
});
