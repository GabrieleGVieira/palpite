import { useEffect } from 'react';

import { connectRealtime, type RealtimeEvent } from '../services/realtime';

type UseRealtimeEventsParams = {
  groupID?: string;
  onEvent: (event: RealtimeEvent) => void;
};

export function useRealtimeEvents({ groupID, onEvent }: UseRealtimeEventsParams) {
  useEffect(() => {
    let cleanup: (() => void) | undefined;
    let isMounted = true;

    connectRealtime({ groupID, onEvent })
      .then((nextCleanup) => {
        if (isMounted) {
          cleanup = nextCleanup;
        } else {
          nextCleanup();
        }
      })
      .catch(() => {
        // REST remains the source of truth when realtime is unavailable.
      });

    return () => {
      isMounted = false;
      cleanup?.();
    };
  }, [groupID, onEvent]);
}
