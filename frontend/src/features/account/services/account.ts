import { uploadAvatar } from './accountAvatarUpload';
import { apiClient } from '../../../shared/services/apiClient';

export type Profile = {
  avatar_url?: string | null;
  display_name: string;
  is_public_profile?: boolean;
};

export async function deleteAccount() {
  await apiClient<Record<string, string>>('/api/v1/me', {
    fallbackError: 'Não foi possível excluir sua conta agora.',
    method: 'DELETE',
  });
}

export async function getProfile() {
  return apiClient<Profile>('/api/v1/me/profile', {
    fallbackError: 'Não foi possível carregar seu perfil.',
  });
}

export async function updateProfile(payload: Profile) {
  return apiClient<Profile>('/api/v1/me/profile', {
    body: JSON.stringify(payload),
    fallbackError: 'Não foi possível atualizar seu perfil.',
    method: 'PATCH',
  });
}

export async function updateProfileAvatar(uri: string, displayName: string) {
  const avatarURL = await uploadAvatar(uri);

  return updateProfile({
    avatar_url: avatarURL,
    display_name: displayName,
  });
}
