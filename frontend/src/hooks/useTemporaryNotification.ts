import { useCallback, useEffect, useState } from 'react';

export function useTemporaryNotification(durationMs = 5000) {
  const [notificationMessage, setNotificationMessage] = useState<string | null>(null);

  const showNotification = useCallback((message: string | null) => {
    if (message) {
      setNotificationMessage(message);
    }
  }, []);

  useEffect(() => {
    if (!notificationMessage) {
      return;
    }

    const timer = setTimeout(() => setNotificationMessage(null), durationMs);
    return () => clearTimeout(timer);
  }, [durationMs, notificationMessage]);

  return {
    notificationMessage,
    showNotification,
  };
}
