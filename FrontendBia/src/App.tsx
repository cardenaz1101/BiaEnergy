import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { BrowserRouter, Navigate, Route, Routes } from 'react-router';
import { AuthProvider } from './auth/AuthContext';
import { RequireAuth } from './auth/RequireAuth';
import { AppLayout } from './components/layout/AppLayout';
import { AnalysisProvider } from './hooks/AnalysisProvider';
import { AnomaliesPage } from './pages/AnomaliesPage';
import { DashboardPage } from './pages/DashboardPage';
import { InvestigationPage } from './pages/InvestigationPage';
import { LoginPage } from './pages/LoginPage';
import { MeterDetailPage } from './pages/MeterDetailPage';
import { MetersPage } from './pages/MetersPage';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { staleTime: 30_000, retry: 1, refetchOnWindowFocus: false },
  },
});

export function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <BrowserRouter>
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route
              element={
                <RequireAuth>
                  <AnalysisProvider>
                    <AppLayout />
                  </AnalysisProvider>
                </RequireAuth>
              }
            >
              <Route index element={<DashboardPage />} />
              <Route path="meters" element={<MetersPage />} />
              <Route path="meters/:meterId" element={<MeterDetailPage />} />
              <Route path="anomalies" element={<AnomaliesPage />} />
              <Route path="anomalies/:anomalyId" element={<InvestigationPage />} />
            </Route>
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </BrowserRouter>
      </AuthProvider>
    </QueryClientProvider>
  );
}
