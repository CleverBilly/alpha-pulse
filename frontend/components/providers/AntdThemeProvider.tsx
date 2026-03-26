"use client";

import type { ReactNode } from "react";
import { App as AntdApp, ConfigProvider, type ThemeConfig, theme } from "antd";

const appTheme: ThemeConfig = {
  algorithm: theme.defaultAlgorithm,
  token: {
    colorPrimary: "#0f766e",
    colorSuccess: "#10b981",
    colorWarning: "#f59e0b",
    colorError: "#ef4444",
    colorInfo: "#0ea5e9",
    colorText: "#0f172a",
    colorTextSecondary: "#64748b",
    colorBorderSecondary: "rgba(15, 23, 42, 0.06)",
    colorBgLayout: "#f0f2f5",
    colorBgContainer: "#ffffff",
    colorFillAlter: "rgba(15, 118, 110, 0.04)",
    fontFamily: "inherit",
    borderRadius: 8,
    borderRadiusLG: 10,
  },
  components: {
    Card: {
      colorBgContainer: "#ffffff",
      headerBg: "transparent",
    },
    Button: {
      controlHeight: 42,
      controlHeightLG: 46,
      borderRadius: 8,
      primaryShadow: "none",
      defaultShadow: "none",
    },
    Menu: {
      itemBg: "transparent",
      horizontalItemSelectedBg: "rgba(15, 118, 110, 0.08)",
      itemSelectedBg: "rgba(15, 118, 110, 0.08)",
      itemSelectedColor: "#0f766e",
      itemColor: "#64748b",
      itemHoverColor: "#0f172a",
      itemHoverBg: "rgba(15, 118, 110, 0.04)",
      activeBarBorderWidth: 0,
      activeBarHeight: 0,
      activeBarWidth: 0,
      horizontalLineHeight: "42px",
      itemBorderRadius: 8,
    },
    Tag: {
      borderRadiusSM: 999,
    },
    Progress: {
      defaultColor: "#0f766e",
      remainingColor: "rgba(15, 118, 110, 0.08)",
    },
  },
};

export default function AntdThemeProvider({ children }: { children: ReactNode }) {
  return (
    <ConfigProvider theme={appTheme}>
      <AntdApp>{children}</AntdApp>
    </ConfigProvider>
  );
}
