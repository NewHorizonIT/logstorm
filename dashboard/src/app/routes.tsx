import { lazy, Suspense } from "react";
import { Navigate, createBrowserRouter } from "react-router";
import { ProtectedRoute } from "@/app/ProtectedRoute";

// Layouts
const MainLayout = lazy(() => import("@/layouts/MainLayout"));
const AuthLayout = lazy(() => import("@/layouts/AuthLayout"));

// Auth Pages
const LoginPage = lazy(() =>
  import("@/modules/auth/pages/LoginPage").then((m) => ({
    default: m.LoginPage,
  }))
);
const RegisterPage = lazy(() =>
  import("@/modules/auth/pages/RegisterPage").then((m) => ({
    default: m.RegisterPage,
  }))
);

// Feature Pages
const Dashboard = lazy(() => import("@/modules/dashboard/pages/Dashboard"));
const LogsExplorerPage = lazy(() =>
  import("@/modules/logs/pages/LogsExplorerPage")
);
const MetricsPage = lazy(() => import("@/modules/metrics/pages/MetricsPage"));
const AlertsPage = lazy(() => import("@/modules/alerts/pages/AlertsPage"));
const ApiKeysPage = lazy(() =>
  import("@/modules/api-keys/pages/ApiKeysPage")
);

// Loading Fallback Component
const LoadingFallback = () => (
  <div className="flex items-center justify-center h-screen bg-[--color-bg]">
    <div className="text-center">
      <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-[--color-primary]"></div>
      <p className="mt-4 text-[--color-text-secondary]">Loading...</p>
    </div>
  </div>
);

const router = createBrowserRouter([
  {
    path: "/",
    element: (
      <Suspense fallback={<LoadingFallback />}>
          <MainLayout />
      </Suspense>
    ),
    children: [
      {
        index: true,
        element: (
          <Suspense fallback={<LoadingFallback />}>
            <Dashboard />
          </Suspense>
        ),
      },
      {
        path: "logs",
        element: (
          <Suspense fallback={<LoadingFallback />}>
            <LogsExplorerPage />
          </Suspense>
        ),
      },
      {
        path: "metrics",
        element: (
          <Suspense fallback={<LoadingFallback />}>
            <MetricsPage />
          </Suspense>
        ),
      },
      {
        path: "alerts",
        element: (
          <Suspense fallback={<LoadingFallback />}>
            <AlertsPage />
          </Suspense>
        ),
      },
      {
        path: "api-keys",
        element: (
          <Suspense fallback={<LoadingFallback />}>
            <ApiKeysPage />
          </Suspense>
        ),
      },
    ],
  },
  {
    path: "/auth",
    element: (
      <Suspense fallback={<LoadingFallback />}>
        <AuthLayout />
      </Suspense>
    ),
    children: [
      {
        path: "login",
        element: (
          <Suspense fallback={<LoadingFallback />}>
            <LoginPage />
          </Suspense>
        ),
      },
      {
        path: "register",
        element: (
          <Suspense fallback={<LoadingFallback />}>
            <RegisterPage />
          </Suspense>
        ),
      },
      {
        index: true,
        element: <Navigate to="login" replace />,
      },
    ],
  },
  {
    path: "*",
    element: <Navigate to="/" replace />,
  },
]);

export default router;
