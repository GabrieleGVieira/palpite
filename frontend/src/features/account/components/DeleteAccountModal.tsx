import { useState } from 'react';
import {
  ActivityIndicator,
  Modal,
  Pressable,
  StyleSheet,
  Text,
  TextInput,
  View,
} from 'react-native';

import { colors } from '../../../shared/theme';

type DeleteAccountModalProps = {
  error: string | null;
  isDeleting: boolean;
  onClose: () => void;
  onConfirm: () => void;
  visible: boolean;
};

const confirmationText = 'EXCLUIR';

export function DeleteAccountModal({
  error,
  isDeleting,
  onClose,
  onConfirm,
  visible,
}: DeleteAccountModalProps) {
  const [confirmation, setConfirmation] = useState('');
  const canConfirm = confirmation.trim().toUpperCase() === confirmationText && !isDeleting;

  function handleClose() {
    if (isDeleting) {
      return;
    }

    setConfirmation('');
    onClose();
  }

  function handleConfirm() {
    if (!canConfirm) {
      return;
    }

    onConfirm();
  }

  return (
    <Modal animationType="fade" onRequestClose={handleClose} transparent visible={visible}>
      <View style={styles.overlay}>
        <View style={styles.dialog}>
          <Text style={styles.title}>Excluir conta</Text>
          <Text style={styles.description}>
            A exclusão da conta é permanente. Seus dados de perfil, nome de exibição, e-mail,
            avatar, preferências e vínculos pessoais com grupos e bolões serão excluídos ou
            anonimizados.
          </Text>
          <Text style={styles.description}>
            Alguns registros técnicos ou históricos podem ser mantidos temporariamente por obrigação
            legal, segurança, prevenção de fraude, resolução de disputas ou integridade histórica de
            bolões, rankings e resultados.
          </Text>
          <Text style={styles.description}>
            Se você for dono de algum grupo, transfira a propriedade para outro Palpiteiro antes
            de excluir sua conta.
          </Text>
          <Text style={styles.description}>O processamento pode levar até 30 dias.</Text>

          <Text style={styles.label}>Digite EXCLUIR para confirmar</Text>
          <TextInput
            autoCapitalize="characters"
            editable={!isDeleting}
            onChangeText={setConfirmation}
            placeholder="EXCLUIR"
            placeholderTextColor="#9ca6a0"
            selectionColor={colors.danger}
            style={styles.input}
            value={confirmation}
          />

          {error ? <Text style={styles.error}>{error}</Text> : null}

          <View style={styles.actions}>
            <Pressable
              disabled={isDeleting}
              onPress={handleClose}
              style={[styles.button, styles.cancelButton, isDeleting && styles.disabled]}>
              <Text style={styles.cancelText}>Cancelar</Text>
            </Pressable>
            <Pressable
              disabled={!canConfirm}
              onPress={handleConfirm}
              style={[styles.button, styles.deleteButton, !canConfirm && styles.disabled]}>
              {isDeleting ? (
                <ActivityIndicator color={colors.white} />
              ) : (
                <Text style={styles.deleteText}>Excluir conta</Text>
              )}
            </Pressable>
          </View>
        </View>
      </View>
    </Modal>
  );
}

const styles = StyleSheet.create({
  overlay: {
    alignItems: 'center',
    backgroundColor: 'rgba(30, 30, 30, 0.44)',
    flex: 1,
    justifyContent: 'center',
    padding: 20,
  },
  dialog: {
    backgroundColor: colors.surface,
    borderRadius: 8,
    maxWidth: 520,
    padding: 20,
    width: '100%',
  },
  title: {
    color: colors.primaryText,
    fontSize: 22,
    fontWeight: '900',
    marginBottom: 10,
  },
  description: {
    color: colors.mutedText,
    fontSize: 14,
    lineHeight: 20,
    marginBottom: 10,
  },
  label: {
    color: colors.primaryText,
    fontSize: 13,
    fontWeight: '800',
    marginBottom: 8,
    marginTop: 4,
  },
  input: {
    backgroundColor: colors.fieldBackground,
    borderColor: colors.fieldBorder,
    borderRadius: 8,
    borderWidth: 1,
    color: colors.primaryText,
    fontSize: 16,
    fontWeight: '800',
    minHeight: 48,
    paddingHorizontal: 14,
  },
  error: {
    color: colors.dangerStrong,
    fontSize: 13,
    fontWeight: '700',
    lineHeight: 18,
    marginTop: 10,
  },
  actions: {
    flexDirection: 'row',
    gap: 10,
    marginTop: 16,
  },
  button: {
    alignItems: 'center',
    borderRadius: 8,
    flex: 1,
    justifyContent: 'center',
    minHeight: 48,
    paddingHorizontal: 12,
  },
  cancelButton: {
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderWidth: 1,
  },
  deleteButton: {
    backgroundColor: colors.dangerStrong,
  },
  cancelText: {
    color: colors.primaryText,
    fontSize: 14,
    fontWeight: '800',
  },
  deleteText: {
    color: colors.white,
    fontSize: 14,
    fontWeight: '800',
  },
  disabled: {
    opacity: 0.58,
  },
});
