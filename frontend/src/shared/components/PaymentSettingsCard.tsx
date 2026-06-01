import { StyleSheet, Text, TextInput, View } from 'react-native';

import { SwitchBox } from './SwitchBox';

type PaymentSettingsCardProps = {
  blockPendingPredictions: boolean;
  isPaid: boolean;
  paymentAmount: string;
  setBlockPendingPredictions: (value: boolean) => void;
  setIsPaid: (value: boolean) => void;
  setPaymentAmount: (value: string) => void;
};

export function PaymentSettingsCard({
  blockPendingPredictions,
  isPaid,
  paymentAmount,
  setBlockPendingPredictions,
  setIsPaid,
  setPaymentAmount,
}: PaymentSettingsCardProps) {
  return (
    <View style={styles.card}>
      <SwitchBox
        title="Participação paga"
        subtitle="Controle manual de quem já pagou"
        value={isPaid}
        onPress={setIsPaid}
      />

      {isPaid ? (
        <>
          <View style={styles.fieldGroup}>
            <Text style={styles.label}>Valor por Palpiteiro</Text>
            <TextInput
              keyboardType="decimal-pad"
              onChangeText={setPaymentAmount}
              placeholder="0,00"
              placeholderTextColor="#8a9a90"
              style={styles.input}
              value={paymentAmount}
            />
          </View>

          <SwitchBox
            title="Bloquear palpites pendentes"
            subtitle="Preparado para uma regra futura"
            value={blockPendingPredictions}
            onPress={setBlockPendingPredictions}
          />
        </>
      ) : null}
    </View>
  );
}

const styles = StyleSheet.create({
  card: {
    backgroundColor: '#ffffff',
    borderColor: '#cfe0c9',
    borderRadius: 8,
    borderWidth: 1,
    gap: 14,
    padding: 16,
  },
  fieldGroup: {
    gap: 8,
  },
  label: {
    color: '#183f2d',
    fontSize: 14,
    fontWeight: '700',
  },
  input: {
    backgroundColor: '#f5f8ef',
    borderColor: '#cfe0c9',
    borderRadius: 8,
    borderWidth: 1,
    color: '#183f2d',
    fontSize: 16,
    minHeight: 52,
    paddingHorizontal: 14,
  },
});
