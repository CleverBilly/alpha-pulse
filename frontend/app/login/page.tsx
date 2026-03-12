"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { authApi } from "@/services/apiClient";

export default function LoginPage() {
  const router = useRouter();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  return (
    <main className="login-screen">
      <section className="login-card">
        <div className="login-card__header">
          <p className="login-card__eyebrow">单用户访问</p>
          <h1 className="login-card__title">登录 Alpha Pulse</h1>
          <p className="login-card__description">登录后才能访问你的交易驾驶舱、图表和信号页面。</p>
        </div>

        <form
          className="login-card__form"
          onSubmit={async (event) => {
            event.preventDefault();
            setSubmitting(true);
            setError(null);
            try {
              await authApi.login(username, password);
              router.replace("/dashboard");
              router.refresh();
            } catch (loginError) {
              setError(loginError instanceof Error ? loginError.message : "登录失败");
            } finally {
              setSubmitting(false);
            }
          }}
        >
          <label className="login-card__field">
            <span>用户名</span>
            <input
              aria-label="用户名"
              value={username}
              onChange={(event) => setUsername(event.target.value)}
              autoComplete="username"
              className="login-card__input"
            />
          </label>

          <label className="login-card__field">
            <span>密码</span>
            <input
              aria-label="密码"
              type="password"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              autoComplete="current-password"
              className="login-card__input"
            />
          </label>

          {error ? <p className="login-card__error">{error}</p> : null}

          <button type="submit" disabled={submitting} className="login-card__submit">
            {submitting ? "登录中..." : "登录"}
          </button>
        </form>
      </section>
    </main>
  );
}
