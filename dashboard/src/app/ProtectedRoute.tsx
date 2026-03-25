import type { ReactNode } from "react";
import { Navigate } from "react-router";
import { useAuthStore } from "@/modules/auth/authStore";

type Props = {
  children: ReactNode;
};

export const ProtectedRoute = ({ children }: Props) => {
  const token = useAuthStore((state) => state.token);

  if (!token) {
    return <Navigate to="/auth/login" replace />;
  }

  return children;
};