import * as FileSystem from 'expo-file-system/legacy';
import { decode } from 'base64-arraybuffer';

import { apiClient } from '../../../shared/services/apiClient';
import { supabase } from '../../../services/supabase';

export type Profile = {
  avatar_url?: string | null;
  display_name: string;
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

export async function uploadAvatar(uri: string) {
  const {
    data: { session },
  } = await supabase.auth.getSession();

  if (!session?.access_token) {
    throw new Error('Sua sessao expirou. Entre novamente.');
  }

  const extension = uri.split('.').pop()?.toLowerCase();
  const normalizedExtension = extension === 'png' || extension === 'webp' ? extension : 'jpg';
  const contentType = contentTypeForExtension(normalizedExtension);
  const base64 = await FileSystem.readAsStringAsync(uri, {
    encoding: FileSystem.EncodingType.Base64,
  });
  const filePath = `${session.user.id}/avatar-${Date.now()}.${normalizedExtension}`;

  const { error } = await supabase.storage.from('avatars').upload(filePath, decode(base64), {
    contentType,
  });

  if (error) {
    throw new Error(error.message || 'Não foi possível enviar a foto.');
  }

  const {
    data: { publicUrl },
  } = supabase.storage.from('avatars').getPublicUrl(filePath);

  return publicUrl;
}

function contentTypeForExtension(extension: string) {
  switch (extension) {
    case 'png':
      return 'image/png';
    case 'webp':
      return 'image/webp';
    default:
      return 'image/jpeg';
  }
}

export async function updateProfileAvatar(uri: string, displayName: string) {
  const avatarURL = await uploadAvatar(uri);

  return updateProfile({
    avatar_url: avatarURL,
    display_name: displayName,
  });
}
