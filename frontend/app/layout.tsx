import type { Metadata } from "next";
import { AntdRegistry } from "@ant-design/nextjs-registry";
import ProAppShell from "@/components/layout/ProAppShell";
import AntdThemeProvider from "@/components/providers/AntdThemeProvider";
import "../styles/globals.css";

export const metadata: Metadata = {
  title: "Alpha Pulse",
  description: "面向个人合约交易者的方向判断与告警工作台",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="zh-CN">
      <body>
        <AntdRegistry>
          <AntdThemeProvider>
            <ProAppShell>{children}</ProAppShell>
          </AntdThemeProvider>
        </AntdRegistry>
      </body>
    </html>
  );
}
