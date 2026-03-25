import { useNavigate } from "react-router";
import { AuthForm } from "../components/AuthForm";
import { useAuth } from "../useAuth";

export const RegisterPage = () => {
  const { register } = useAuth();
  const navigate = useNavigate();

  const handleRegister = async (email: string, password: string) => {
    await register(email, password);
    navigate("/auth/login");
  };

  return (
    
      <AuthForm type="register" onSubmit={handleRegister} />
  );
};