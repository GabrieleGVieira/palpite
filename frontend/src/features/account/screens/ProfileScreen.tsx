import { useEffect, useState } from 'react';
import { Alert, Image, Platform, ScrollView, StyleSheet, Text, TextInput, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { StatusBar } from 'expo-status-bar';
import * as ImagePicker from 'expo-image-picker';

import { BackButton } from '../../../shared/components/BackButton';
import { FinishButton } from '../../../shared/components/FinishButton';
import { LoadingIndicator } from '../../../shared/components/LoadingIndicator';
import { SwitchBox } from '../../../shared/components/SwitchBox';
import { colors } from '../../../shared/theme';
import { getProfile, updateProfile, updateProfileAvatar } from '../services/account';

type Props = {
  fallbackName?: string;
  onBack: () => void;
};

export function ProfileScreen({ fallbackName, onBack }: Props) {
  const [displayName, setDisplayName] = useState(fallbackName ?? '');
  const [avatarURL, setAvatarURL] = useState('');
  const [isPublicProfile, setIsPublicProfile] = useState(true);
  const [isLoading, setIsLoading] = useState(true);
  const [isSaving, setIsSaving] = useState(false);
  const [isUploadingAvatar, setIsUploadingAvatar] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function loadProfile() {
      setIsLoading(true);
      setError(null);

      try {
        const profile = await getProfile();
        setDisplayName(profile.display_name || fallbackName || '');
        setAvatarURL(profile.avatar_url ?? '');
        setIsPublicProfile(profile.is_public_profile ?? true);
      } catch (loadError) {
        setError(
          loadError instanceof Error ? loadError.message : 'Não foi possível carregar seu perfil.',
        );
      } finally {
        setIsLoading(false);
      }
    }

    void loadProfile();
  }, [fallbackName]);

  async function handleSave() {
    setIsSaving(true);
    setError(null);

    try {
      const profile = await updateProfile({
        avatar_url: avatarURL.trim() || null,
        display_name: displayName.trim(),
        is_public_profile: isPublicProfile,
      });
      setDisplayName(profile.display_name);
      setAvatarURL(profile.avatar_url ?? '');
      setIsPublicProfile(profile.is_public_profile ?? true);
      Alert.alert('Perfil atualizado', 'Suas informações foram salvas.');
    } catch (saveError) {
      setError(
        saveError instanceof Error ? saveError.message : 'Não foi possível atualizar seu perfil.',
      );
    } finally {
      setIsSaving(false);
    }
  }

  async function handlePickAvatar() {
    setError(null);
    const normalizedDisplayName = displayName.trim();
    if (!normalizedDisplayName) {
      setError('Informe seu nome antes de enviar a foto.');
      return;
    }

    if (Platform.OS !== 'web') {
      const permission = await ImagePicker.requestMediaLibraryPermissionsAsync();
      if (!permission.granted) {
        setError('Permita acesso as fotos para escolher uma imagem.');
        return;
      }
    }

    const result = await ImagePicker.launchImageLibraryAsync({
      allowsEditing: true,
      aspect: [1, 1],
      mediaTypes: ImagePicker.MediaTypeOptions.Images,
      quality: 0.82,
    });

    if (result.canceled || !result.assets[0]?.uri) {
      return;
    }

    setIsUploadingAvatar(true);
    try {
      const profile = await updateProfileAvatar(result.assets[0].uri, normalizedDisplayName);
      setDisplayName(profile.display_name || displayName);
      setAvatarURL(profile.avatar_url ?? '');
    } catch (uploadError) {
      setError(
        uploadError instanceof Error ? uploadError.message : 'Não foi possível enviar a foto.',
      );
    } finally {
      setIsUploadingAvatar(false);
    }
  }

  return (
    <SafeAreaView style={styles.safeArea}>
      <StatusBar style="dark" />
      <ScrollView contentContainerStyle={styles.container} showsVerticalScrollIndicator={false}>
        <BackButton onPress={onBack} />

        <View>
          <Text style={styles.title}>Editar perfil</Text>
          <Text style={styles.subtitle}>
            Nome e foto aparecem para Palpiteiros dos seus grupos.
          </Text>
        </View>

        {isLoading ? <LoadingIndicator text="Carregando perfil..." /> : null}

        {!isLoading ? (
          <View style={styles.formCard}>
            {avatarURL.trim() ? (
              <Image source={{ uri: avatarURL.trim() }} style={styles.avatar} />
            ) : (
              <View style={styles.avatarFallback}>
                <Text style={styles.avatarText}>{initials(displayName)}</Text>
              </View>
            )}

            <FinishButton
              isLoading={isUploadingAvatar}
              loadingLabel="Enviando foto..."
              onPress={handlePickAvatar}
              waitingLabel={avatarURL.trim() ? 'Trocar foto' : 'Escolher foto'}
            />

            <View style={styles.field}>
              <Text style={styles.label}>Nome</Text>
              <TextInput
                autoCapitalize="words"
                onChangeText={setDisplayName}
                placeholder="Seu nome"
                style={styles.input}
                value={displayName}
              />
            </View>

            <View style={styles.publicProfileBox}>
              <SwitchBox
                onPress={setIsPublicProfile}
                subtitle="Permite aparecer na busca de adicionar amigos."
                title="Perfil público"
                value={isPublicProfile}
              />
            </View>

            {error ? <Text style={styles.errorText}>{error}</Text> : null}

            <FinishButton
              isLoading={isSaving}
              loadingLabel="Salvando..."
              onPress={handleSave}
              waitingLabel="Salvar perfil"
            />
          </View>
        ) : null}
      </ScrollView>
    </SafeAreaView>
  );
}

function initials(name: string) {
  const trimmed = name.trim();
  if (!trimmed) {
    return '?';
  }

  return trimmed
    .split(/\s+/)
    .slice(0, 2)
    .map((part) => part[0])
    .join('')
    .toUpperCase();
}

const styles = StyleSheet.create({
  safeArea: {
    backgroundColor: colors.background,
    flex: 1,
  },
  container: {
    backgroundColor: colors.background,
    flexGrow: 1,
    gap: 18,
    paddingHorizontal: 24,
    paddingVertical: 32,
  },
  title: {
    color: colors.primaryText,
    fontSize: 30,
    fontWeight: '800',
  },
  subtitle: {
    color: colors.mutedText,
    fontSize: 15,
    lineHeight: 22,
    marginTop: 6,
  },
  formCard: {
    alignItems: 'center',
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: 8,
    borderWidth: 1,
    gap: 14,
    padding: 18,
  },
  avatar: {
    borderRadius: 44,
    height: 88,
    width: 88,
  },
  avatarFallback: {
    alignItems: 'center',
    backgroundColor: '#e3efe0',
    borderRadius: 44,
    height: 88,
    justifyContent: 'center',
    width: 88,
  },
  avatarText: {
    color: colors.primary,
    fontSize: 28,
    fontWeight: '800',
  },
  field: {
    alignSelf: 'stretch',
    gap: 6,
  },
  label: {
    color: colors.primaryText,
    fontSize: 13,
    fontWeight: '800',
  },
  input: {
    backgroundColor: '#ffffff',
    borderColor: colors.border,
    borderRadius: 8,
    borderWidth: 1,
    color: colors.primaryText,
    fontSize: 15,
    minHeight: 48,
    paddingHorizontal: 12,
  },
  errorText: {
    alignSelf: 'stretch',
    color: colors.danger,
    fontSize: 13,
    lineHeight: 18,
  },
  publicProfileBox: {
    alignSelf: 'stretch',
  },
});
