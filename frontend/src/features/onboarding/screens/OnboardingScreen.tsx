import { StatusBar } from 'expo-status-bar';
import { useState } from 'react';
import { Image, Pressable, StyleSheet, Text, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

type OnboardingScreenProps = {
  onFinish: () => void;
};

const steps = [
  {
    eyebrow: 'Bolão com amigos',
    title: 'Crie grupos e dispute rodada por rodada',
    description:
      'Monte bolões privados, convide sua turma e acompanhe quem sobe no ranking dos Palpiteiros.',
    image: require('../../../../assets/onboarding-groups.png'),
  },
  {
    eyebrow: 'PalpitAI',
    title: 'Receba contexto antes de palpitar',
    description:
      'Veja tendências, momento das seleções e sinais úteis da PalpitAI como apoio complementar.',
    image: require('../../../../assets/onboarding-ai.png'),
  },
  {
    eyebrow: 'Tempo real',
    title: 'Ranking vivo durante os jogos',
    description:
      'Acompanhe mudanças na pontuação, palpites do grupo e resultados sem perder o clima de torcida.',
    image: require('../../../../assets/onboarding-live.png'),
  },
];

export function OnboardingScreen({ onFinish }: OnboardingScreenProps) {
  const [currentStep, setCurrentStep] = useState(0);
  const step = steps[currentStep];
  const isLastStep = currentStep === steps.length - 1;

  function handleNext() {
    if (isLastStep) {
      onFinish();
      return;
    }

    setCurrentStep((previousStep) => previousStep + 1);
  }

  return (
    <SafeAreaView style={styles.safeArea}>
      <StatusBar style="dark" />
      <View style={styles.container}>
        <View style={styles.pitchMarkTop} />
        <View style={styles.pitchCircle} />

        <View style={styles.topBar}>
          <View style={styles.logoMark}>
            <Image
              accessibilityIgnoresInvertColors
              resizeMode="cover"
              source={require('../../../../assets/splash-palpite.png')}
              style={styles.logoImage}
            />
          </View>

          <Pressable onPress={onFinish} style={styles.skipButton}>
            <Text style={styles.skipButtonText}>Pular</Text>
          </Pressable>
        </View>

        <View style={styles.hero}>
          <View style={styles.imageFrame}>
            <Image
              accessibilityIgnoresInvertColors
              resizeMode="cover"
              source={step.image}
              style={styles.onboardingImage}
            />
          </View>
        </View>

        <View style={styles.content}>
          <Text style={styles.eyebrow}>{step.eyebrow}</Text>
          <Text style={styles.title}>{step.title}</Text>
          <Text style={styles.description}>{step.description}</Text>
        </View>

        <View style={styles.footer}>
          <View style={styles.dots}>
            {steps.map((item, index) => (
              <View
                key={item.title}
                style={[styles.dot, index === currentStep ? styles.dotActive : null]}
              />
            ))}
          </View>

          <Pressable onPress={handleNext} style={styles.primaryButton}>
            <Text style={styles.primaryButtonText}>{isLastStep ? 'Começar' : 'Próximo'}</Text>
          </Pressable>
        </View>
      </View>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safeArea: {
    flex: 1,
    backgroundColor: '#f5f8ef',
  },
  container: {
    flex: 1,
    backgroundColor: '#f5f8ef',
    justifyContent: 'space-between',
    paddingHorizontal: 24,
    paddingVertical: 32,
  },
  pitchMarkTop: {
    borderColor: 'rgba(255, 255, 255, 0.72)',
    borderRadius: 8,
    borderWidth: 2,
    height: 116,
    left: 24,
    position: 'absolute',
    right: 24,
    top: -42,
  },
  pitchCircle: {
    borderColor: 'rgba(32, 111, 67, 0.12)',
    borderRadius: 150,
    borderWidth: 2,
    height: 300,
    position: 'absolute',
    right: -136,
    top: 120,
    width: 300,
  },
  topBar: {
    alignItems: 'center',
    flexDirection: 'row',
    justifyContent: 'space-between',
  },
  logoMark: {
    alignItems: 'center',
    backgroundColor: '#ffffff',
    borderColor: '#d9e7d4',
    borderRadius: 28,
    borderWidth: 1,
    height: 56,
    justifyContent: 'center',
    overflow: 'hidden',
    shadowColor: '#1e5c39',
    shadowOffset: { height: 8, width: 0 },
    shadowOpacity: 0.12,
    shadowRadius: 16,
    width: 56,
  },
  logoImage: {
    height: 68,
    transform: [{ scale: 1.16 }],
    width: 68,
  },
  skipButton: {
    paddingHorizontal: 4,
    paddingVertical: 8,
  },
  skipButtonText: {
    color: '#1f7a4a',
    fontSize: 14,
    fontWeight: '800',
  },
  hero: {
    alignItems: 'center',
    justifyContent: 'center',
    marginVertical: 28,
  },
  imageFrame: {
    aspectRatio: 1.08,
    backgroundColor: '#ffffff',
    borderColor: '#d9e7d4',
    borderRadius: 8,
    borderWidth: 1,
    maxHeight: 330,
    overflow: 'hidden',
    shadowColor: '#1e5c39',
    shadowOffset: { height: 14, width: 0 },
    shadowOpacity: 0.18,
    shadowRadius: 24,
    width: '100%',
  },
  onboardingImage: {
    height: '100%',
    width: '100%',
  },
  content: {
    minHeight: 170,
  },
  eyebrow: {
    color: '#1f7a4a',
    fontSize: 13,
    fontWeight: '900',
    letterSpacing: 0,
    marginBottom: 10,
    textTransform: 'uppercase',
  },
  title: {
    color: '#123d2a',
    fontSize: 34,
    fontWeight: '900',
    letterSpacing: 0,
    lineHeight: 40,
  },
  description: {
    color: '#486654',
    fontSize: 16,
    lineHeight: 24,
    marginTop: 14,
  },
  footer: {
    gap: 24,
  },
  dots: {
    flexDirection: 'row',
    gap: 8,
  },
  dot: {
    backgroundColor: '#cfe0c9',
    borderRadius: 4,
    height: 8,
    width: 8,
  },
  dotActive: {
    backgroundColor: '#1f7a4a',
    width: 28,
  },
  primaryButton: {
    alignItems: 'center',
    backgroundColor: '#1f7a4a',
    borderRadius: 8,
    justifyContent: 'center',
    minHeight: 54,
  },
  primaryButtonText: {
    color: '#ffffff',
    fontSize: 16,
    fontWeight: '800',
  },
});
