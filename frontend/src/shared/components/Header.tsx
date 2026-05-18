import { Image, StyleSheet, Text, View } from 'react-native';

type HeaderProps = {
  title: string;
  subtitle: string;
  topSpacing?: number;
  withImage?: boolean;
};

export function Header({ title, subtitle, topSpacing = 24, withImage = true }: HeaderProps) {
  return (
    <View style={[styles.header, { paddingTop: topSpacing }]}>
      {withImage && (
        <View style={styles.logoMark}>
          <Image
            accessibilityIgnoresInvertColors
            resizeMode="cover"
            source={require('../../../assets/splash-palpitai.png')}
            style={styles.logoImage}
          />
        </View>
      )}
      <Text style={styles.title}>{title}</Text>
      <Text style={styles.subtitle}>{subtitle}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  header: {
    paddingTop: 24,
  },
  logoMark: {
    alignItems: 'center',
    backgroundColor: '#ffffff',
    borderColor: '#d9e7d4',
    borderRadius: 32,
    borderWidth: 1,
    height: 64,
    justifyContent: 'center',
    marginBottom: 24,
    overflow: 'hidden',
    shadowColor: '#1e5c39',
    shadowOffset: { height: 8, width: 0 },
    shadowOpacity: 0.12,
    shadowRadius: 16,
    width: 64,
  },
  logoImage: {
    height: 76,
    transform: [{ scale: 1.18 }],
    width: 76,
  },
  title: {
    color: '#123d2a',
    fontSize: 36,
    fontWeight: '800',
    letterSpacing: 0,
  },
  subtitle: {
    color: '#486654',
    fontSize: 16,
    lineHeight: 24,
    marginTop: 12,
    maxWidth: 360,
  },
});
