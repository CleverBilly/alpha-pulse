"use client";

import type { ReactNode } from "react";
import { App as AntdApp, ConfigProvider, type ThemeConfig, theme } from "antd";

const appTheme: ThemeConfig = {
  algorithm: theme.defaultAlgorithm,
  token: {
    colorPrimary: "#0f766e",
    colorSuccess: "#15803d",
    colorWarning: "#c2410c",
    colorError: "#be123c",
    colorInfo: "#0369a1",
    colorText: "#172033",
    colorTextSecondary: "#52607a",
    colorBorderSecondary: "rgba(23, 32, 51, 0.08)",
    colorBgLayout: "transparent",
    colorBgContainer: "rgba(255, 253, 248, 0.92)",
    colorFillAlter: "rgba(15, 118, 110, 0.06)",
    fontFamily: "\"Azeret Mono\", \"IBM Plex Sans\", \"Noto Sans SC\", sans-serif",
    borderRadius: 18,
    borderRadiusLG: 24,
    boxShadow: "0 18px 60px rgba(32, 42, 63, 0.10)",
    boxShadowSecondary: "0 10px 30px rgba(32, 42, 63, 0.08)",
  },
  components: {
    Layout: {
      bodyBg: "transparent",
      headerBg: "transparent",
      siderBg: "transparent",
      triggerBg: "#0f766e",
    },
    Card: {
      colorBgContainer: "rgba(255, 253, 248, 0.88)",
      headerBg: "transparent",
    },
    Button: {
      controlHeight: 42,
      controlHeightLG: 46,
      borderRadius: 16,
      primaryShadow: "none",
      defaultShadow: "none",
    },
    Menu: {
      itemBg: "transparent",
      horizontalItemSelectedBg: "rgba(15, 118, 110, 0.10)",
      itemSelectedBg: "rgba(15, 118, 110, 0.10)",
      itemSelectedColor: "#0f766e",
      itemColor: "#52607a",
      itemHoverColor: "#172033",
      activeBarBorderWidth: 0,
      horizontalLineHeight: "42px",
      itemBorderRadius: 14,
    },
    Tag: {
      borderRadiusSM: 999,
    },
    Progress: {
      defaultColor: "#0f766e",
      remainingColor: "rgba(15, 118, 110, 0.10)",
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
