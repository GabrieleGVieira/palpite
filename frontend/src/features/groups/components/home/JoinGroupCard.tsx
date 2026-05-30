import { StyleSheet, Text, TextInput, Pressable, View } from 'react-native';

type JoinGroupCardProps = {
  inviteCode: string;
  setInviteCode: (value: string) => void;
  onJoinGroup: () => void;
  isJoiningGroup: boolean;
};

export function JoinGroupCard({
  inviteCode,
  setInviteCode,
  onJoinGroup,
  isJoiningGroup,
}: JoinGroupCardProps) {
  return (
    <View style={styles.card}>
      <View>
        <Text style={styles.title}>Entrar em um grupo</Text>
        <Text style={styles.subtitle}>Use o código de convite recebido.</Text>
      </View>

      <View style={styles.form}>
        <TextInput
          autoCapitalize="characters"
          onChangeText={setInviteCode}
          placeholder="CÓDIGO"
          placeholderTextColor="#7c8898"
          style={styles.input}
          value={inviteCode}
        />
        <Pressable
          disabled={isJoiningGroup}
          onPress={onJoinGroup}
          style={[styles.button, isJoiningGroup && styles.buttonDisabled]}>
          <Text style={styles.buttonText}>{isJoiningGroup ? 'Entrando...' : 'Entrar'}</Text>
        </Pressable>
      </View>
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
  title: {
    color: '#123d2a',
    fontSize: 18,
    fontWeight: '800',
  },
  subtitle: {
    color: '#486654',
    fontSize: 13,
    marginTop: 4,
  },
  form: {
    flexDirection: 'row',
    gap: 10,
  },
  input: {
    backgroundColor: '#f5f8ef',
    borderColor: '#cfe0c9',
    borderRadius: 8,
    borderWidth: 1,
    color: '#183f2d',
    flex: 1,
    fontSize: 16,
    fontWeight: '800',
    minHeight: 48,
    paddingHorizontal: 14,
  },
  button: {
    alignItems: 'center',
    backgroundColor: '#1f7a4a',
    borderRadius: 8,
    justifyContent: 'center',
    minHeight: 48,
    paddingHorizontal: 18,
  },
  buttonText: {
    color: '#ffffff',
    fontSize: 14,
    fontWeight: '800',
  },
  buttonDisabled: {
    opacity: 0.72,
  },
});
