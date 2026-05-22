import { useState } from "react";
import { Link } from "react-router";

type Props = {
  type: "login" | "register";
  onSubmit: (email: string, password: string) => Promise<void> | void;
};

export const AuthForm = ({ type, onSubmit }: Props) => {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const isLogin = type === "login";

  const handleSubmit = async () => {
    if (!email.trim() || !password.trim() || isSubmitting) {
      return;
    }

    try {
      setIsSubmitting(true);
      setError(null);
      await onSubmit(email.trim(), password);
    } catch {
      setError(
        isLogin
          ? "Unable to sign in. Please check your credentials."
          : "Unable to create account. Please try again.",
      );
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="w-full text-white">
      <header>
        <h2 className="font-display text-3xl leading-tight text-white">
          {isLogin ? "Welcome back" : "Create your account"}
        </h2>
        <p className="mt-2 text-sm text-slate-400">
          {isLogin
            ? "Sign in to access your observability dashboard"
            : "Start monitoring your systems in minutes"}
        </p>
      </header>

      <form
        className="mt-8 space-y-4"
        onSubmit={(e) => {
          e.preventDefault();
          handleSubmit();
        }}
      >
        <div>
          <label className="mb-1 block text-sm font-medium text-slate-300">Email</label>
          <input
            className="h-11 w-full rounded-xl border border-slate-700 bg-slate-900/80 px-3 text-sm text-white outline-none transition focus:border-cyan-400/80 focus:ring-2 focus:ring-cyan-400/20"
            placeholder="you@company.com"
            autoComplete="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
          />
        </div>

        <div>
          <div className="mb-1 flex items-center justify-between">
            <label className="block text-sm font-medium text-slate-300">Password</label>
            {isLogin && (
              <button
                type="button"
                className="text-xs font-medium text-cyan-300 transition hover:text-cyan-200"
              >
                Forgot?
              </button>
            )}
          </div>

          <div className="relative">
            <input
              className="h-11 w-full rounded-xl border border-slate-700 bg-slate-900/80 px-3 pr-20 text-sm text-white outline-none transition focus:border-cyan-400/80 focus:ring-2 focus:ring-cyan-400/20"
              type={showPassword ? "text" : "password"}
              placeholder="At least 8 characters"
              autoComplete={isLogin ? "current-password" : "new-password"}
              value={password}
              onChange={(e) => setPassword(e.target.value)}
            />

            <button
              type="button"
              className="absolute right-2 top-1/2 -translate-y-1/2 rounded-md px-2 py-1 text-xs font-medium text-slate-300 transition hover:bg-slate-800 hover:text-white"
              onClick={() => setShowPassword((prev) => !prev)}
            >
              {showPassword ? "Hide" : "Show"}
            </button>
          </div>
        </div>

        {error && (
          <p className="rounded-lg border border-rose-400/35 bg-rose-400/10 px-3 py-2 text-sm text-rose-200">
            {error}
          </p>
        )}

        <button
          type="submit"
          className="h-11 w-full rounded-xl bg-linear-to-r from-cyan-500 to-blue-500 text-sm font-semibold text-white shadow-[0_10px_25px_-12px_rgba(6,182,212,0.9)] transition enabled:hover:brightness-110 enabled:active:translate-y-px disabled:cursor-not-allowed disabled:opacity-45"
          disabled={!email.trim() || !password.trim() || isSubmitting}
        >
          {isSubmitting ? "Please wait..." : isLogin ? "Sign in" : "Create account"}
        </button>
      </form>

      <div className="mt-6 border-t border-slate-800 pt-5 text-sm text-slate-400">
        {isLogin ? (
          <>
            New to LogStorm?{" "}
            <Link className="font-medium text-cyan-300 transition hover:text-cyan-200" to="/auth/register">
              Create an account
            </Link>
          </>
        ) : (
          <>
            Already have an account?{" "}
            <Link className="font-medium text-cyan-300 transition hover:text-cyan-200" to="/auth/login">
              Sign in
            </Link>
          </>
        )}
      </div>

      <div className="mt-4 text-xs text-slate-500">
        By continuing, you agree to LogStorm terms and privacy policy.
      </div>
    </div>
  );
};