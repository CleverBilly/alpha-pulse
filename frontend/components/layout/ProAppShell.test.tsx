import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

const navigationState = vi.hoisted(() => ({
  pathname: "/dashboard",
}));

vi.mock("next/navigation", () => ({
  usePathname: () => navigationState.pathname,
}));

vi.mock("@/components/alerts/AlertCenter", () => ({
  default: () => <button type="button">告警</button>,
}));

vi.mock("./SignalStatusBadge", () => ({
  default: () => <span>BUY · 82%</span>,
}));

import ProAppShell from "./ProAppShell";

describe("ProAppShell", () => {
  beforeEach(() => {
    window.localStorage.clear();
    navigationState.pathname = "/dashboard";
  });

  it("renders the shell, active navigation item, and children", () => {
    render(
      <ProAppShell>
        <div>content</div>
      </ProAppShell>,
    );

    expect(screen.getByTestId("command-center-shell")).toHaveAttribute("data-collapsed", "false");
    expect(screen.getByTestId("command-center-shell")).toHaveAttribute("data-shell-style", "integrated");
    expect(screen.getByTestId("command-center-rail")).toBeInTheDocument();
    expect(screen.getByTestId("command-center-mobile-topbar")).toBeInTheDocument();
    expect(screen.getByTestId("command-center-canvas")).toHaveAttribute("data-shell-surface", "continuous");
    expect(screen.getByRole("link", { name: /驾驶舱/i })).toHaveAttribute("data-active", "true");
    expect(screen.getByRole("link", { name: /图表/i })).toHaveAttribute("data-active", "false");
    expect(screen.getByText("content")).toBeInTheDocument();
  });

  it("toggles collapsed state and persists it to localStorage", async () => {
    const user = userEvent.setup();

    render(
      <ProAppShell>
        <div>x</div>
      </ProAppShell>,
    );

    const shell = screen.getByTestId("command-center-shell");
    expect(shell).toHaveAttribute("data-collapsed", "false");

    await user.click(screen.getByRole("button", { name: "收起侧边栏" }));

    expect(shell).toHaveAttribute("data-collapsed", "true");
    expect(window.localStorage.getItem("sidebar-collapsed")).toBe("true");
  });

  it("renders the sidebar dock actions", () => {
    render(
      <ProAppShell>
        <div>content</div>
      </ProAppShell>,
    );

    expect(screen.getByTestId("command-center-shell-actions")).toContainElement(
      screen.getByRole("button", { name: "收起侧边栏" }),
    );
    const dock = screen.getByTestId("command-center-dock");
    expect(dock).toBeInTheDocument();
    expect(within(dock).getByText("BUY · 82%")).toBeInTheDocument();
    expect(within(dock).getByRole("button", { name: "告警" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "退出登录" })).toBeInTheDocument();
  });

  it("opens and closes the mobile drawer navigation", async () => {
    const user = userEvent.setup();

    render(
      <ProAppShell>
        <div>content</div>
      </ProAppShell>,
    );

    const shell = screen.getByTestId("command-center-shell");
    const drawer = screen.getByTestId("command-center-mobile-drawer");

    expect(shell).toHaveAttribute("data-mobile-nav-open", "false");
    expect(drawer).toHaveAttribute("hidden");

    await user.click(screen.getByRole("button", { name: "打开导航菜单" }));

    expect(shell).toHaveAttribute("data-mobile-nav-open", "true");
    expect(drawer).not.toHaveAttribute("hidden");
    expect(screen.getAllByRole("link", { name: /自动交易/i })).toHaveLength(2);

    await user.click(screen.getByRole("button", { name: "关闭导航菜单" }));

    expect(shell).toHaveAttribute("data-mobile-nav-open", "false");
    expect(drawer).toHaveAttribute("hidden");
  });

  it("bypasses the shell on the login route", () => {
    navigationState.pathname = "/login";

    render(
      <ProAppShell>
        <div>login-content</div>
      </ProAppShell>,
    );

    expect(screen.getByText("login-content")).toBeInTheDocument();
    expect(screen.queryByTestId("command-center-shell")).not.toBeInTheDocument();
    expect(screen.queryByRole("link", { name: /驾驶舱/i })).not.toBeInTheDocument();
  });
});
