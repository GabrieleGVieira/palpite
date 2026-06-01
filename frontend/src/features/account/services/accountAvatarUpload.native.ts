import { decode } from 'base64-arraybuffer';
import * as FileSystem from 'expo-file-system/legacy';

import { supabase } from '../../../services/supabase';

export async function uploadAvatar(uri: string) {
  const {
    data: { session },
  } = await supabase.auth.getSession();

  if (!session?.access_token) {
    throw new Error('Sua sessao expirou. Entre novamente.');
  }

  const extension = uri.split('.').pop()?.toLowerCase();
  const normalizedExtension = normalizeImageExtension(extension);
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

function normalizeImageExtension(extension?: string) {
  return extension === 'png' || extension === 'webp' ? extension : 'jpg';
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
