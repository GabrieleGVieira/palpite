import { useState } from 'react';
import { Modal, Pressable, ScrollView, StyleSheet, Text, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useQuery } from '@tanstack/react-query';

import { BackButton } from '../../../shared/components/BackButton';
import { EmptyBox } from '../../../shared/components/EmptyBox';
import { LoadingIndicator } from '../../../shared/components/LoadingIndicator';
import { colors } from '../../../shared/theme';
import {
  listPalpicoinRanking,
  listWalletTransactions,
  PALPICOIN_NOTICE,
  getWallet,
} from '../services/palpicoins';

type Props = {
  onBack: () => void;
};

export function PalpicoinsScreen({ onBack }: Props) {
  const [activeTab, setActiveTab] = useState<'wallet' | 'ranking'>('wallet');
  const [isRulesVisible, setIsRulesVisible] = useState(false);
  const walletQuery = useQuery({ queryFn: getWallet, queryKey: ['me', 'wallet'] });
  const transactionsQuery = useQuery({
    queryFn: listWalletTransactions,
    queryKey: ['me', 'wallet', 'transactions'],
  });
  const rankingQuery = useQuery({
    enabled: activeTab === 'ranking',
    queryFn: listPalpicoinRanking,
    queryKey: ['rankings', 'palpicoins'],
  });

  return (
    <SafeAreaView style={styles.safeArea}>
      <ScrollView contentContainerStyle={styles.container} showsVerticalScrollIndicator={false}>
        <View style={styles.topBar}>
          <BackButton onPress={onBack} />
          <Text style={styles.notice}>{PALPICOIN_NOTICE}</Text>
        </View>

        <Text style={styles.title}>Palpicoins</Text>

        <View style={styles.summary}>
          <Text style={styles.summaryLabel}>Saldo atual</Text>
          <Text style={styles.balance}>{walletQuery.data?.balance ?? '...'}</Text>
          <View style={styles.summaryRow}>
            <Text style={styles.summaryText}>Ganhos: {walletQuery.data?.totalEarned ?? 0}</Text>
            <Text style={styles.summaryText}>Gastos: {walletQuery.data?.totalSpent ?? 0}</Text>
          </View>
        </View>

        <Pressable onPress={() => setIsRulesVisible(true)} style={styles.primaryButton}>
          <Text style={styles.primaryButtonText}>Como ganhar Palpicoins</Text>
        </Pressable>

        <View style={styles.tabs}>
          <Pressable
            onPress={() => setActiveTab('wallet')}
            style={[styles.tab, activeTab === 'wallet' && styles.tabActive]}>
            <Text style={[styles.tabText, activeTab === 'wallet' && styles.tabTextActive]}>
              Histórico
            </Text>
          </Pressable>
          <Pressable
            onPress={() => setActiveTab('ranking')}
            style={[styles.tab, activeTab === 'ranking' && styles.tabActive]}>
            <Text style={[styles.tabText, activeTab === 'ranking' && styles.tabTextActive]}>
              Ranking global
            </Text>
          </Pressable>
        </View>

        {activeTab === 'wallet' && transactionsQuery.isLoading ? (
          <LoadingIndicator text="Carregando histórico..." />
        ) : null}
        {activeTab === 'wallet' && !transactionsQuery.isLoading && !transactionsQuery.data?.items.length ? (
          <EmptyBox title="Sem movimentações" text="Seu histórico aparecerá aqui." />
        ) : null}
        {activeTab === 'wallet'
          ? transactionsQuery.data?.items.map((item) => (
              <View key={item.id} style={styles.rowCard}>
                <View style={styles.rowTextBlock}>
                  <Text style={styles.rowTitle}>{item.description}</Text>
                  <Text style={styles.rowSubtitle}>{new Date(item.createdAt).toLocaleString()}</Text>
                </View>
                <Text style={[styles.amount, item.amount < 0 && styles.amountNegative]}>
                  {item.amount > 0 ? '+' : ''}
                  {item.amount}
                </Text>
              </View>
            ))
          : null}

        {activeTab === 'ranking' && rankingQuery.isLoading ? (
          <LoadingIndicator text="Carregando ranking..." />
        ) : null}
        {activeTab === 'ranking'
          ? rankingQuery.data?.ranking.map((entry) => (
              <View
                key={entry.userId}
                style={[styles.rowCard, entry.isCurrentUser && styles.currentRankingCard]}>
                <Text style={styles.position}>{entry.posicao}</Text>
                <View style={styles.rowTextBlock}>
                  <Text style={styles.rowTitle}>{entry.nome}</Text>
                  <Text style={styles.rowSubtitle}>{entry.isCurrentUser ? 'Você' : 'Palpiteiro'}</Text>
                </View>
                <Text style={styles.amount}>{entry.saldo}</Text>
              </View>
            ))
          : null}
      </ScrollView>

      <Modal animationType="slide" transparent visible={isRulesVisible}>
        <View style={styles.modalBackdrop}>
          <View style={styles.modalContent}>
            <Text style={styles.modalTitle}>Como ganhar Palpicoins</Text>
            <Text style={styles.modalText}>Criar conta: +1000</Text>
            <Text style={styles.modalText}>Acertar vencedor: +10</Text>
            <Text style={styles.modalText}>Acertar empate: +15</Text>
            <Text style={styles.modalText}>Acertar placar exato: +50</Text>
            <Text style={styles.modalNotice}>{PALPICOIN_NOTICE}</Text>
            <Pressable onPress={() => setIsRulesVisible(false)} style={styles.primaryButton}>
              <Text style={styles.primaryButtonText}>Fechar</Text>
            </Pressable>
          </View>
        </View>
      </Modal>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safeArea: { backgroundColor: colors.background, flex: 1 },
  container: { gap: 16, paddingHorizontal: 24, paddingVertical: 32 },
  topBar: { alignItems: 'center', flexDirection: 'row', gap: 12 },
  notice: { color: colors.mutedText, flex: 1, fontSize: 12, lineHeight: 17 },
  title: { color: colors.primaryDark, fontSize: 32, fontWeight: '900' },
  summary: {
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: 8,
    borderWidth: 1,
    padding: 18,
  },
  summaryLabel: { color: colors.mutedText, fontSize: 12, fontWeight: '800' },
  balance: { color: colors.primaryDark, fontSize: 42, fontWeight: '900', marginTop: 4 },
  summaryRow: { flexDirection: 'row', gap: 14, marginTop: 10 },
  summaryText: { color: colors.mutedText, fontSize: 13, fontWeight: '700' },
  primaryButton: {
    alignItems: 'center',
    backgroundColor: colors.primary,
    borderRadius: 8,
    minHeight: 48,
    justifyContent: 'center',
    paddingHorizontal: 14,
  },
  primaryButtonText: { color: colors.white, fontSize: 14, fontWeight: '900' },
  tabs: { backgroundColor: '#edf3e8', borderRadius: 8, flexDirection: 'row', gap: 6, padding: 6 },
  tab: { alignItems: 'center', borderRadius: 7, flex: 1, minHeight: 42, justifyContent: 'center' },
  tabActive: { backgroundColor: colors.surface },
  tabText: { color: colors.mutedText, fontSize: 13, fontWeight: '900' },
  tabTextActive: { color: colors.primary },
  rowCard: {
    alignItems: 'center',
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: 8,
    borderWidth: 1,
    flexDirection: 'row',
    gap: 12,
    padding: 14,
  },
  currentRankingCard: { borderColor: colors.primary, borderWidth: 2 },
  rowTextBlock: { flex: 1 },
  rowTitle: { color: colors.primaryDark, fontSize: 14, fontWeight: '900' },
  rowSubtitle: { color: colors.mutedText, fontSize: 12, marginTop: 4 },
  amount: { color: colors.primary, fontSize: 16, fontWeight: '900' },
  amountNegative: { color: colors.danger },
  position: { color: colors.primaryDark, fontSize: 18, fontWeight: '900', width: 30 },
  modalBackdrop: {
    backgroundColor: 'rgba(18, 61, 42, 0.28)',
    flex: 1,
    justifyContent: 'flex-end',
  },
  modalContent: {
    backgroundColor: colors.surface,
    borderTopLeftRadius: 8,
    borderTopRightRadius: 8,
    gap: 10,
    padding: 24,
  },
  modalTitle: { color: colors.primaryDark, fontSize: 22, fontWeight: '900' },
  modalText: { color: colors.text, fontSize: 15, fontWeight: '700' },
  modalNotice: { color: colors.mutedText, fontSize: 12, lineHeight: 18, marginVertical: 6 },
});
