import { QueryClientProvider } from '@tanstack/react-query';

import { AuthProvider } from './src/features/auth/store/AuthProvider';
import { AppNavigator } from './src/navigation/AppNavigator';
import { queryClient } from './src/shared/query/queryClient';

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <AppNavigator />
      </AuthProvider>
    </QueryClientProvider>
  );
}
