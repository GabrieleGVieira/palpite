import { supabase } from '../../../services/supabase';

export async function uploadAvatar(uri: string) {
  const {
    data: { session },
  } = await supabase.auth.getSession();

  if (!session?.access_token) {
    throw new Error('Sua sessao expirou. Entre novamente.');
  }

  const response = await fetch(uri);
  const blob = await response.blob();
  const contentType = blob.type || 'image/jpeg';
  const extension = extensionForContentType(contentType);
  const filePath = `${session.user.id}/avatar-${Date.now()}.${extension}`;

  const { error } = await supabase.storage.from('avatars').upload(filePath, blob, {
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

function extensionForContentType(contentType: string) {
  switch (contentType) {
    case 'image/png':
      return 'png';
    case 'image/webp':
      return 'webp';
    default:
      return 'jpg';
  }
}
