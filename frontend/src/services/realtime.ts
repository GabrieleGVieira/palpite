import { supabase } from './supabase';

const apiURL = process.env.EXPO_PUBLIC_API_URL ?? 'http://localhost:3000';

export type RealtimeEvent = {
  name: 'match.updated' | 'match.finished' | 'match.goal' | 'ranking.updated' | string;
  payload: Record<string, unknown>;
  room?: string;
};

type RealtimeOptions = {
  groupID?: string;
  onEvent: (event: RealtimeEvent) => void;
};

export async function connectRealtime({ groupID, onEvent }: RealtimeOptions) {
  const {
    data: { session },
  } = await supabase.auth.getSession();

  if (!session?.access_token) {
    throw new Error('Sua sessao expirou. Entre novamente para receber atualizacoes ao vivo.');
  }

  const accessToken = session.access_token;
  let isClosed = false;
  let socket: WebSocket | null = null;
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null;

  function connect() {
    const url = new URL('/ws', websocketURL(apiURL));
    url.searchParams.set('token', accessToken);
    if (groupID) {
      url.searchParams.set('group_id', groupID);
    }

    socket = new WebSocket(url.toString());

    socket.onmessage = (message) => {
      try {
        onEvent(JSON.parse(String(message.data)) as RealtimeEvent);
      } catch {
        // Ignore malformed realtime messages. The REST API remains the source of truth.
      }
    };

    socket.onclose = () => {
      if (!isClosed) {
        reconnectTimer = setTimeout(connect, 2000);
      }
    };

    socket.onerror = () => {
      socket?.close();
    };
  }

  connect();

  return () => {
    isClosed = true;
    if (reconnectTimer) {
      clearTimeout(reconnectTimer);
    }
    socket?.close();
  };
}

function websocketURL(value: string) {
  if (value.startsWith('https://')) {
    return value.replace('https://', 'wss://');
  }

  return value.replace('http://', 'ws://');
}
