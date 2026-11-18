"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Eye, EyeSlash, CircleNotch } from "@phosphor-icons/react";

import { loginSchema, type LoginInput } from "@/lib/validations/auth";
import { loginApi } from "@/lib/api/auth";
import { ApiError } from "@/lib/api/client";
import { useAuthStore } from "@/store/auth";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { cn } from "@/lib/utils";

export function LoginForm() {
  const router = useRouter();
  const setAuth = useAuthStore((s) => s.setAuth);
  const [showPassword, setShowPassword] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<LoginInput>({
    resolver: zodResolver(loginSchema),
  });

  async function onSubmit(data: LoginInput) {
    setFormError(null);
    try {
      const { data: tokenData } = await loginApi(data);
      setAuth(tokenData.access_token, {
        id: "",
        email: data.email,
        fullName: "",
      });
      router.push("/projects");
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        setFormError("Invalid email or password.");
      } else {
        setFormError("Something went wrong. Please try again.");
      }
    }
  }

  return (
    <div className="rounded-card border-border-card bg-surface w-full max-w-sm border p-8">
      <h1 className="text-text-primary mb-1 text-xl font-semibold">Sign in</h1>
      <p className="text-text-muted mb-6 text-sm">Welcome back. Enter your details to continue.</p>

      <form onSubmit={(e) => void handleSubmit(onSubmit)(e)} className="flex flex-col gap-4">
        {/* Email */}
        <div className="flex flex-col gap-1.5">
          <Label htmlFor="email">Email</Label>
          <Input
            id="email"
            type="email"
            placeholder="you@company.com"
            autoComplete="email"
            aria-invalid={!!errors.email}
            className={cn(
              "rounded-input border-border-input bg-bg-base text-text-primary placeholder:text-text-disabled focus-visible:border-accent-focus focus-visible:ring-accent-focus/20 h-9 px-3 text-sm",
              errors.email && "border-status-danger",
            )}
            {...register("email")}
          />
          {errors.email && <p className="text-status-danger text-xs">{errors.email.message}</p>}
        </div>

        {/* Password */}
        <div className="flex flex-col gap-1.5">
          <Label htmlFor="password">Password</Label>
          <div className="relative">
            <Input
              id="password"
              type={showPassword ? "text" : "password"}
              placeholder="At least 8 characters"
              autoComplete="current-password"
              aria-invalid={!!errors.password}
              className={cn(
                "rounded-input border-border-input bg-bg-base text-text-primary placeholder:text-text-disabled focus-visible:border-accent-focus focus-visible:ring-accent-focus/20 h-9 px-3 pr-9 text-sm",
                errors.password && "border-status-danger",
              )}
              {...register("password")}
            />
            <button
              type="button"
              onClick={() => setShowPassword((v) => !v)}
              className="text-text-muted hover:text-text-primary absolute top-1/2 right-3 -translate-y-1/2"
              tabIndex={-1}
            >
              {showPassword ? <EyeSlash size={15} /> : <Eye size={15} />}
            </button>
          </div>
          {errors.password && (
            <p className="text-status-danger text-xs">{errors.password.message}</p>
          )}
        </div>

        {/* Form error */}
        {formError && (
          <p className="rounded-control border-status-danger/30 bg-status-danger-bg text-status-danger border px-3 py-2 text-xs">
            {formError}
          </p>
        )}

        <button
          type="submit"
          disabled={isSubmitting}
          className="rounded-input bg-accent-brand mt-1 flex h-9 w-full items-center justify-center gap-2 text-sm font-medium text-white transition-opacity hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-60"
        >
          {isSubmitting && <CircleNotch size={15} className="animate-spin" />}
          Sign in
        </button>
      </form>

      <p className="text-text-muted mt-6 text-center text-sm">
        Don&apos;t have an account?{" "}
        <Link href="/register" className="text-accent-brand hover:underline">
          Sign up
        </Link>
      </p>
    </div>
  );
}
