"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Eye, EyeSlash, CircleNotch } from "@phosphor-icons/react";

import { registerSchema, type RegisterInput } from "@/lib/validations/auth";
import { registerApi, loginApi } from "@/lib/api/auth";
import { ApiError } from "@/lib/api/client";
import { useAuthStore } from "@/store/auth";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { cn } from "@/lib/utils";

export function RegisterForm() {
  const router = useRouter();
  const setAuth = useAuthStore((s) => s.setAuth);
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirm, setShowConfirm] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<RegisterInput>({
    resolver: zodResolver(registerSchema),
  });

  async function onSubmit(data: RegisterInput) {
    setFormError(null);

    try {
      await registerApi({
        full_name: data.full_name,
        email: data.email,
        password: data.password,
      });
    } catch (err) {
      if (err instanceof ApiError && err.status === 409) {
        setFormError("An account with this email already exists.");
      } else {
        setFormError("Something went wrong. Please try again.");
      }
      return;
    }

    try {
      const tokenRes = await loginApi({ email: data.email, password: data.password });
      setAuth(tokenRes.access_token, {
        id: "",
        email: data.email,
        fullName: data.full_name,
      });
      router.push("/projects");
    } catch {
      // Account created but auto-login failed — redirect to login
      router.push("/login");
    }
  }

  return (
    <div className="rounded-card border-border-card bg-surface w-full max-w-sm border p-8">
      <h1 className="text-text-primary mb-1 text-xl font-semibold">Create your account</h1>
      <p className="text-text-muted mb-6 text-sm">
        Start collecting and searching your logs in minutes.
      </p>

      <form onSubmit={(e) => void handleSubmit(onSubmit)(e)} className="flex flex-col gap-4">
        {/* Full name */}
        <div className="flex flex-col gap-1.5">
          <Label htmlFor="full_name">Full name</Label>
          <Input
            id="full_name"
            type="text"
            placeholder="Your name"
            autoComplete="name"
            aria-invalid={!!errors.full_name}
            className={cn(
              "rounded-input border-border-input bg-bg-base text-text-primary placeholder:text-text-disabled focus-visible:border-accent-focus focus-visible:ring-accent-focus/20 h-9 px-3 text-sm",
              errors.full_name && "border-status-danger",
            )}
            {...register("full_name")}
          />
          {errors.full_name && (
            <p className="text-status-danger text-xs">{errors.full_name.message}</p>
          )}
        </div>

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
              autoComplete="new-password"
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

        {/* Confirm password */}
        <div className="flex flex-col gap-1.5">
          <Label htmlFor="confirm_password">Confirm Password</Label>
          <div className="relative">
            <Input
              id="confirm_password"
              type={showConfirm ? "text" : "password"}
              placeholder="Re-enter your password"
              autoComplete="new-password"
              aria-invalid={!!errors.confirm_password}
              className={cn(
                "rounded-input border-border-input bg-bg-base text-text-primary placeholder:text-text-disabled focus-visible:border-accent-focus focus-visible:ring-accent-focus/20 h-9 px-3 pr-9 text-sm",
                errors.confirm_password && "border-status-danger",
              )}
              {...register("confirm_password")}
            />
            <button
              type="button"
              onClick={() => setShowConfirm((v) => !v)}
              className="text-text-muted hover:text-text-primary absolute top-1/2 right-3 -translate-y-1/2"
              tabIndex={-1}
            >
              {showConfirm ? <EyeSlash size={15} /> : <Eye size={15} />}
            </button>
          </div>
          {errors.confirm_password && (
            <p className="text-status-danger text-xs">{errors.confirm_password.message}</p>
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
          Create Account
        </button>
      </form>

      <p className="text-text-muted mt-6 text-center text-sm">
        Already have an account?{" "}
        <Link href="/login" className="text-accent-brand hover:underline">
          Sign in
        </Link>
      </p>
    </div>
  );
}
