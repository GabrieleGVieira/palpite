import { useState } from 'react';
import { Pressable, StyleSheet, Text, TextInput, View } from 'react-native';

import { EmptyBox } from '../../../../shared/components/EmptyBox';
import { LoadingIndicator } from '../../../../shared/components/LoadingIndicator';
import type { GroupPayment, GroupPaymentsSummary, PaymentStatus } from '../../services/groups';

type GroupAdminPaymentsProps = {
  isLoadingPayments: boolean;
  loadPayments: () => void;
  onUpdatePayment: (
    payment: GroupPayment,
    status: PaymentStatus,
    amountPaid: number,
    amountExpected: number,
    paymentMethod: string,
    notes: string,
  ) => void;
  payments: GroupPayment[];
  summary: GroupPaymentsSummary | null;
  updatingPaymentUserID: string | null;
};

const statusLabels: Record<PaymentStatus, string> = {
  exempt: 'Isento',
  paid: 'Pago',
  pending: 'Pendente',
  refunded: 'Reembolsado',
};

export function GroupAdminPayments({
  isLoadingPayments,
  loadPayments,
  onUpdatePayment,
  payments,
  summary,
  updatingPaymentUserID,
}: GroupAdminPaymentsProps) {
  const pendingPayments = payments.filter((payment) => payment.status === 'pending');

  return (
    <View style={styles.card}>
      <View style={styles.header}>
        <View>
          <Text style={styles.cardTitle}>Pagamentos pendentes</Text>
        </View>
        <Pressable onPress={loadPayments} style={styles.refreshButton}>
          <Text style={styles.refreshButtonText}>Atualizar</Text>
        </Pressable>
      </View>

      <View style={styles.summaryGrid}>
        <SummaryCard label="Arrecadado" value={formatMoney(summary?.total_paid ?? 0)} />
        <SummaryCard label="Pendente" value={formatMoney(summary?.total_pending ?? 0)} />
        <SummaryCard label="Pagos" value={String(summary?.paid_count ?? 0)} />
        <SummaryCard label="Pendentes" value={String(summary?.pending_count ?? 0)} />
      </View>

      {isLoadingPayments ? <LoadingIndicator text="Carregando pagamentos..." /> : null}

      {!isLoadingPayments && pendingPayments.length === 0 ? (
        <EmptyBox
          title="Nenhum pagamento pendente."
          text="Pagamentos confirmados aparecem na lista de membros."
        />
      ) : null}

      {pendingPayments.map((payment) => (
        <PaymentItem
          key={payment.user_id}
          payment={payment}
          isUpdating={updatingPaymentUserID === payment.user_id}
          onUpdatePayment={onUpdatePayment}
        />
      ))}
    </View>
  );
}

function SummaryCard({ label, value }: { label: string; value: string }) {
  return (
    <View style={styles.summaryCard}>
      <Text style={styles.summaryValue}>{value}</Text>
      <Text style={styles.summaryLabel}>{label}</Text>
    </View>
  );
}

function PaymentItem({
  isUpdating,
  onUpdatePayment,
  payment,
}: {
  isUpdating: boolean;
  onUpdatePayment: GroupAdminPaymentsProps['onUpdatePayment'];
  payment: GroupPayment;
}) {
  const [isOpen, setIsOpen] = useState(false);
  const [amountPaid, setAmountPaid] = useState(
    payment.amount_paid > 0 ? String(payment.amount_paid) : '',
  );
  const [paymentMethod, setPaymentMethod] = useState(payment.payment_method);
  const [notes, setNotes] = useState(payment.notes);
  const [fieldError, setFieldError] = useState<string | null>(null);

  function submit(status: PaymentStatus) {
    setFieldError(null);

    if (status === 'paid' && amountPaid.trim() === '') {
      setFieldError('Informe o valor pago.');
      return;
    }

    const parsedAmount = Number(amountPaid.replace(',', '.'));
    if (status === 'paid' && (!Number.isFinite(parsedAmount) || parsedAmount <= 0)) {
      setFieldError('Informe um valor pago maior que zero.');
      return;
    }

    const nextAmount = Number.isFinite(parsedAmount) ? parsedAmount : 0;
    onUpdatePayment(
      payment,
      status,
      nextAmount,
      status === 'exempt' ? 0 : payment.amount_expected,
      paymentMethod,
      notes,
    );
  }

  return (
    <View style={styles.paymentRow}>
      <Pressable onPress={() => setIsOpen((current) => !current)} style={styles.paymentSummary}>
        <View style={styles.memberInfo}>
          <View style={styles.summaryTitleRow}>
            <Text style={styles.memberName}>
              {payment.display_name || `Usuário ${payment.user_id.slice(0, 8)}`}
            </Text>
            <Text style={styles.chevron}>{isOpen ? '▲' : '▼'}</Text>
          </View>

          <View style={styles.paymentFacts}>
            <PaymentFact label="Esperado" value={formatMoney(payment.amount_expected)} />
            <PaymentFact label="Pago" value={formatMoney(payment.amount_paid)} />
          </View>
        </View>

        <View style={[styles.statusBadge, styles[`status_${payment.status}`]]}>
          <Text style={[styles.statusText, styles[`statusText_${payment.status}`]]}>
            {statusLabels[payment.status]}
          </Text>
        </View>
      </Pressable>

      {isOpen ? (
        <View style={styles.dropdownContent}>
          <View style={styles.expectedBox}>
            <Text style={styles.label}>Valor esperado</Text>
            <Text style={styles.expectedValue}>{formatMoney(payment.amount_expected)}</Text>
          </View>

          <View style={styles.field}>
            <Text style={styles.label}>Valor pago</Text>
            <TextInput
              keyboardType="decimal-pad"
              onChangeText={(value) => {
                setAmountPaid(value);
                if (fieldError) {
                  setFieldError(null);
                }
              }}
              placeholder="0,00"
              placeholderTextColor="#8a9a90"
              style={[styles.input, fieldError && styles.inputError]}
              value={amountPaid}
            />
            {fieldError ? <Text style={styles.fieldError}>{fieldError}</Text> : null}
          </View>

          <View style={styles.field}>
            <Text style={styles.label}>Método</Text>
            <TextInput
              onChangeText={setPaymentMethod}
              placeholder="Pix, dinheiro..."
              placeholderTextColor="#8a9a90"
              style={styles.input}
              value={paymentMethod}
            />
          </View>

          <View style={styles.field}>
            <Text style={styles.label}>Observações</Text>
            <TextInput
              multiline
              onChangeText={setNotes}
              placeholder="Opcional"
              placeholderTextColor="#8a9a90"
              style={[styles.input, styles.notesInput]}
              value={notes}
            />
          </View>

          <View style={styles.actions}>
            <ActionButton
              disabled={isUpdating}
              label="Pago"
              onPress={() => submit('paid')}
              variant="primary"
            />
            <ActionButton
              disabled={isUpdating}
              label="Isentar"
              onPress={() => submit('exempt')}
              variant="secondary"
            />
          </View>
        </View>
      ) : null}
    </View>
  );
}

function PaymentFact({ label, value }: { label: string; value: string }) {
  return (
    <View style={styles.paymentFact}>
      <Text style={styles.paymentFactLabel}>{label}</Text>
      <Text style={styles.paymentFactValue}>{value}</Text>
    </View>
  );
}

function ActionButton({
  disabled,
  label,
  onPress,
  variant,
}: {
  disabled: boolean;
  label: string;
  onPress: () => void;
  variant: 'primary' | 'secondary';
}) {
  return (
    <Pressable
      disabled={disabled}
      onPress={onPress}
      style={[
        styles.actionButton,
        variant === 'primary' ? styles.actionButtonPrimary : styles.actionButtonSecondary,
        disabled && styles.disabled,
      ]}>
      <Text
        style={[
          styles.actionButtonText,
          variant === 'primary'
            ? styles.actionButtonPrimaryText
            : styles.actionButtonSecondaryText,
        ]}>
        {disabled ? 'Salvando...' : label}
      </Text>
    </Pressable>
  );
}

function formatMoney(value: number) {
  return new Intl.NumberFormat('pt-BR', {
    currency: 'BRL',
    style: 'currency',
  }).format(value);
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
  summaryGrid: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 8,
  },
  summaryCard: {
    backgroundColor: '#f5f8ef',
    borderColor: '#d8e6d1',
    borderRadius: 8,
    borderWidth: 1,
    flexBasis: '48%',
    flexGrow: 1,
    padding: 12,
  },
  summaryValue: {
    color: '#123d2a',
    fontSize: 16,
    fontWeight: '800',
  },
  summaryLabel: {
    color: '#486654',
    fontSize: 12,
    marginTop: 4,
  },
  paymentRow: {
    borderTopColor: '#edf3e8',
    borderTopWidth: 1,
    paddingTop: 14,
  },
  paymentSummary: {
    alignItems: 'center',
    flexDirection: 'row',
    gap: 10,
    justifyContent: 'space-between',
  },
  summaryTitleRow: {
    alignItems: 'center',
    flexDirection: 'row',
    gap: 8,
  },
  chevron: {
    color: '#1f7a4a',
    fontSize: 20,
    fontWeight: '800',
    lineHeight: 22,
  },
  paymentFacts: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 8,
    marginTop: 8,
  },
  paymentFact: {
    borderRadius: 8,
    paddingHorizontal: 10,
    paddingVertical: 8,
  },
  paymentFactLabel: {
    color: '#486654',
    fontSize: 11,
    fontWeight: '700',
  },
  paymentFactValue: {
    color: '#183f2d',
    fontSize: 13,
    fontWeight: '800',
    marginTop: 2,
  },
  memberInfo: {
    flex: 1,
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
  statusBadge: {
    borderRadius: 8,
    paddingHorizontal: 10,
    paddingVertical: 7,
  },
  statusText: {
    fontSize: 12,
    fontWeight: '800',
  },
  status_pending: {
    backgroundColor: '#fff7dd',
  },
  status_paid: {
    backgroundColor: '#e4f6ea',
  },
  status_exempt: {
    backgroundColor: '#edf3e8',
  },
  status_refunded: {
    backgroundColor: '#fde9e7',
  },
  statusText_pending: {
    color: '#8a5d00',
  },
  statusText_paid: {
    color: '#1f7a4a',
  },
  statusText_exempt: {
    color: '#486654',
  },
  statusText_refunded: {
    color: '#b23b32',
  },
  dropdownContent: {
    gap: 12,
    paddingTop: 14,
  },
  expectedBox: {
    backgroundColor: '#f5f8ef',
    borderColor: '#cfe0c9',
    borderRadius: 8,
    borderWidth: 1,
    gap: 4,
    padding: 12,
  },
  expectedValue: {
    color: '#123d2a',
    fontSize: 16,
    fontWeight: '800',
  },
  field: {
    flex: 1,
    gap: 6,
  },
  label: {
    color: '#183f2d',
    fontSize: 12,
    fontWeight: '700',
  },
  input: {
    backgroundColor: '#f5f8ef',
    borderColor: '#cfe0c9',
    borderRadius: 8,
    borderWidth: 1,
    color: '#183f2d',
    fontSize: 14,
    minHeight: 46,
    paddingHorizontal: 12,
  },
  inputError: {
    borderColor: '#c23f34',
  },
  fieldError: {
    color: '#c23f34',
    fontSize: 12,
    fontWeight: '700',
  },
  notesInput: {
    minHeight: 68,
    paddingTop: 10,
  },
  actions: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 8,
  },
  actionButton: {
    borderRadius: 8,
    borderWidth: 1,
    flexGrow: 1,
    paddingHorizontal: 10,
    paddingVertical: 10,
  },
  actionButtonPrimary: {
    backgroundColor: '#1f7a4a',
    borderColor: '#1f7a4a',
  },
  actionButtonSecondary: {
    backgroundColor: '#f5f8ef',
    borderColor: '#cfe0c9',
  },
  actionButtonText: {
    fontSize: 12,
    fontWeight: '800',
    textAlign: 'center',
  },
  actionButtonPrimaryText: {
    color: '#ffffff',
  },
  actionButtonSecondaryText: {
    color: '#1f7a4a',
  },
  disabled: {
    opacity: 0.7,
  },
});
