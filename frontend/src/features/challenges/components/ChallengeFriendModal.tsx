import { useState } from 'react';
import { Alert, Modal, Pressable, StyleSheet, Text, TextInput, View } from 'react-native';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

import { listFriends } from '../../friends/services/friends';
import { colors } from '../../../shared/theme';
import { PALPICOIN_NOTICE } from '../../palpicoins/services/palpicoins';
import { createChallenge } from '../services/challenges';

type Props = {
  matchID: string | null;
  onClose: () => void;
};

export function ChallengeFriendModal({ matchID, onClose }: Props) {
  const [selectedFriendID, setSelectedFriendID] = useState<string | null>(null);
  const [stakeAmount, setStakeAmount] = useState('100');
  const queryClient = useQueryClient();
  const friendsQuery = useQuery({
    enabled: Boolean(matchID),
    queryFn: listFriends,
    queryKey: ['friends'],
  });
  const mutation = useMutation({
    mutationFn: createChallenge,
    onError: (error) => {
      Alert.alert('Erro', error instanceof Error ? error.message : 'Não foi possível criar o desafio.');
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['challenges'] });
      await queryClient.invalidateQueries({ queryKey: ['me', 'wallet'] });
      Alert.alert('Desafio enviado', 'Seu amigo recebeu o desafio.');
      onClose();
    },
  });

  function submit() {
    if (!matchID || !selectedFriendID || Number(stakeAmount) <= 0) {
      Alert.alert('Atenção', 'Selecione um amigo e informe um valor.');
      return;
    }
    mutation.mutate({
      matchId: matchID,
      opponentId: selectedFriendID,
      stakeAmount: Number(stakeAmount),
    });
  }

  return (
    <Modal animationType="slide" transparent visible={Boolean(matchID)}>
      <View style={styles.backdrop}>
        <View style={styles.content}>
          <Text style={styles.title}>Desafiar amigo</Text>
          <Text style={styles.notice}>{PALPICOIN_NOTICE}</Text>

          <View style={styles.friendList}>
            {friendsQuery.data?.map((friend) => (
              <Pressable
                key={friend.userId}
                onPress={() => setSelectedFriendID(friend.userId)}
                style={[styles.friendRow, selectedFriendID === friend.userId && styles.friendRowActive]}>
                <Text
                  style={[
                    styles.friendName,
                    selectedFriendID === friend.userId && styles.friendNameActive,
                  ]}>
                  {friend.name}
                </Text>
              </Pressable>
            ))}
          </View>

          <TextInput
            keyboardType="number-pad"
            onChangeText={setStakeAmount}
            placeholder="Valor"
            style={styles.input}
            value={stakeAmount}
          />

          <View style={styles.actions}>
            <Pressable onPress={onClose} style={styles.secondaryButton}>
              <Text style={styles.secondaryButtonText}>Cancelar</Text>
            </Pressable>
            <Pressable disabled={mutation.isPending} onPress={submit} style={styles.primaryButton}>
              <Text style={styles.primaryButtonText}>
                {mutation.isPending ? 'Enviando...' : 'Confirmar'}
              </Text>
            </Pressable>
          </View>
        </View>
      </View>
    </Modal>
  );
}

const styles = StyleSheet.create({
  backdrop: {
    backgroundColor: 'rgba(18, 61, 42, 0.28)',
    flex: 1,
    justifyContent: 'flex-end',
  },
  content: {
    backgroundColor: colors.surface,
    borderTopLeftRadius: 8,
    borderTopRightRadius: 8,
    gap: 12,
    maxHeight: '82%',
    padding: 24,
  },
  title: { color: colors.primaryDark, fontSize: 22, fontWeight: '900' },
  notice: { color: colors.mutedText, fontSize: 12, lineHeight: 18 },
  friendList: { gap: 8, maxHeight: 260 },
  friendRow: {
    borderColor: colors.border,
    borderRadius: 8,
    borderWidth: 1,
    paddingHorizontal: 14,
    paddingVertical: 12,
  },
  friendRowActive: { backgroundColor: colors.primary, borderColor: colors.primary },
  friendName: { color: colors.primaryDark, fontSize: 14, fontWeight: '900' },
  friendNameActive: { color: colors.white },
  input: {
    backgroundColor: colors.fieldBackground,
    borderColor: colors.fieldBorder,
    borderRadius: 8,
    borderWidth: 1,
    color: colors.text,
    fontSize: 18,
    fontWeight: '900',
    minHeight: 48,
    paddingHorizontal: 14,
  },
  actions: { flexDirection: 'row', gap: 10 },
  primaryButton: {
    alignItems: 'center',
    backgroundColor: colors.primary,
    borderRadius: 8,
    flex: 1,
    justifyContent: 'center',
    minHeight: 48,
  },
  primaryButtonText: { color: colors.white, fontSize: 14, fontWeight: '900' },
  secondaryButton: {
    alignItems: 'center',
    borderColor: colors.border,
    borderRadius: 8,
    borderWidth: 1,
    flex: 1,
    justifyContent: 'center',
    minHeight: 48,
  },
  secondaryButtonText: { color: colors.primary, fontSize: 14, fontWeight: '900' },
});
