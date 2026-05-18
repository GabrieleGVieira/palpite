import { StyleSheet, Text, View } from 'react-native';
import type { JoinRequest } from '../../services/groups';
import { FinishButton } from '../../../../shared/components/FinishButton';

type GroupAdminRequestItemProps = {
  request: JoinRequest;
  isApproving: boolean;
  onApprove: (request: JoinRequest) => void;
};

export function GroupAdminRequestItem({
  request,
  isApproving,
  onApprove,
}: GroupAdminRequestItemProps) {
  return (
    <View style={styles.requestRow}>
      <View style={styles.requestInfo}>
        <Text style={styles.requestUser}>
          {request.display_name || `Usuário ${request.user_id.slice(0, 8)}`}
        </Text>
        <Text style={styles.requestMeta}>Solicitou entrada</Text>
      </View>
      <FinishButton
        isLoading={isApproving}
        onPress={() => onApprove(request)}
        loadingLabel="Aprovando..."
        waitingLabel="Aprovar"
      />
    </View>
  );
}

const styles = StyleSheet.create({
  requestRow: {
    alignItems: 'center',
    borderTopColor: '#edf3e8',
    borderTopWidth: 1,
    flexDirection: 'row',
    gap: 10,
    justifyContent: 'space-between',
    paddingTop: 12,
  },
  requestInfo: {
    flex: 1,
  },
  requestUser: {
    color: '#183f2d',
    fontSize: 14,
    fontWeight: '800',
  },
  requestMeta: {
    color: '#486654',
    fontSize: 12,
    marginTop: 3,
  },
  approveButton: {
    alignItems: 'center',
    backgroundColor: '#1f7a4a',
    borderRadius: 8,
    justifyContent: 'center',
    minHeight: 40,
    paddingHorizontal: 12,
  },
  approveButtonText: {
    color: '#ffffff',
    fontSize: 13,
    fontWeight: '800',
  },
  buttonDisabled: {
    opacity: 0.72,
  },
});
