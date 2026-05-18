import { Pressable, StyleSheet, Text, View } from 'react-native';

type FooterProps = {
  question: string;
  buttonLabel: string;
  onButtonPress: () => void;
};

export function Footer({ question, buttonLabel, onButtonPress }: FooterProps) {
  return (
    <View style={styles.container}>
      <Text style={styles.question}>{question}</Text>
      <Pressable onPress={onButtonPress} style={styles.button}>
        <Text style={styles.buttonText}>{buttonLabel}</Text>
      </Pressable>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    alignItems: 'center',
    marginTop: 28,
  },
  question: {
    color: '#425b4f',
    fontSize: 14,
    marginBottom: 12,
  },
  button: {
    backgroundColor: '#1b5b38',
    borderRadius: 7,
    paddingVertical: 7,
    paddingHorizontal: 10,
  },
  buttonText: {
    color: '#fff',
    fontSize: 10,
    fontWeight: '700',
    textAlign: 'center',
  },
});
