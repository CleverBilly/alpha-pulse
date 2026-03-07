import type { Metadata } from "next";
import Link from "next/link";
import "../styles/globals.css";

export const metadata: Metadata = {
  title: "Alpha Pulse",
  description: "AI Crypto Trading Dashboard",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="zh-CN">
      <body className="bg-bg text-text">
        <div className="min-h-screen">
          <header className="border-b border-slate-200 bg-white/90 backdrop-blur">
            <div className="mx-auto flex max-w-7xl items-center justify-between px-4 py-4">
              <h1 className="text-xl font-bold tracking-wide">Alpha Pulse</h1>
              <nav className="flex items-center gap-6 text-sm font-medium text-slate-700">
                <Link href="/dashboard" className="hover:text-accent">
                  Dashboard
                </Link>
                <Link href="/chart" className="hover:text-accent">
                  Chart
                </Link>
                <Link href="/signals" className="hover:text-accent">
                  Signals
                </Link>
                <Link href="/market" className="hover:text-accent">
                  Market
                </Link>
              </nav>
            </div>
          </header>
          <main className="mx-auto max-w-7xl px-4 py-6">{children}</main>
        </div>
      </body>
    </html>
  );
}
