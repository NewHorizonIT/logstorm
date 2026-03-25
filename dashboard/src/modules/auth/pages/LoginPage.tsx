import { useNavigate } from "react-router";
import { AuthForm } from "../components/AuthForm";
import { useAuth } from "../useAuth";

export const LoginPage = () => {
  const { login } = useAuth();
  const navigate = useNavigate();

  const handleLogin = async (email: string, password: string) => {
    await login(email, password);
    navigate("/");
  };

  return (
    <>
      <AuthForm type="login" onSubmit={handleLogin} />
    </>
  );
};