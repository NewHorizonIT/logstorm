import { Outlet } from "react-router";

const AuthLayout = () => {
  return (
    <div className="relative min-h-screen overflow-hidden bg-bg text-text-primary">
      <div className="pointer-events-none absolute -top-30 -left-30 h-90 w-90 rounded-full bg-cyan-400/12 blur-3xl" />
      <div className="pointer-events-none absolute -bottom-36 right-0 h-96 w-96 rounded-full bg-orange-400/12 blur-3xl" />

      <div className="relative mx-auto flex min-h-screen w-full max-w-6xl items-center px-6 py-10 lg:px-10">
        <div className="grid w-full gap-6 lg:grid-cols-[1.1fr_0.9fr]">
          <section className="auth-fade rounded-3xl border border-white/10 bg-white/2 p-8 backdrop-blur-sm lg:p-10">
            <p className="mb-4 inline-flex items-center rounded-full border border-cyan-300/25 bg-cyan-300/10 px-3 py-1 text-xs font-medium uppercase tracking-[0.12em] text-cyan-200">
              Real-time observability
            </p>
            <h1 className="font-display text-4xl leading-tight text-white lg:text-5xl">
              Keep your systems loud.
              <br />
              Keep outages quiet.
            </h1>
            <p className="mt-5 max-w-xl text-base leading-relaxed text-slate-300">
              LogStorm unifies ingestion, stream processing, anomaly alerts, and dashboards into one workflow so your team can detect and resolve incidents faster.
            </p>

            <div className="mt-9 grid gap-3 sm:grid-cols-2">
              <article className="rounded-2xl border border-white/10 bg-slate-900/50 p-4">
                <p className="text-xs uppercase tracking-widest text-slate-400">Throughput</p>
                <p className="mt-2 font-display text-2xl text-white">4.2M events/min</p>
              </article>
              <article className="rounded-2xl border border-white/10 bg-slate-900/50 p-4">
                <p className="text-xs uppercase tracking-widest text-slate-400">Processing lag</p>
                <p className="mt-2 font-display text-2xl text-white">&lt; 180ms</p>
              </article>
            </div>
          </section>

          <section className="auth-fade auth-fade-delay rounded-3xl border border-white/12 bg-card p-8 shadow-[0_25px_70px_-20px_rgba(5,7,20,0.75)] lg:p-10">
            <div className="mb-8">
              <div className="font-display text-3xl font-semibold text-white">LogStorm</div>
              <p className="mt-1 text-sm text-slate-400">Security-first access for your telemetry workspace</p>
            </div>

            <Outlet />
          </section>
        </div>
      </div>
    </div>
  );
};

export default AuthLayout;
