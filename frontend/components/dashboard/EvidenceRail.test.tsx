import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import EvidenceRail from "@/components/dashboard/EvidenceRail";
import { useMarketStore } from "@/store/marketStore";
import { buildMockMarketStoreState } from "../../test/fixtures/marketStore";

vi.mock("next/link", () => ({
  default: ({ children, href }: { children: React.ReactNode; href: string }) => <a href={href}>{children}</a>,
}));

vi.mock("@/store/marketStore", () => ({
  useMarketStore: vi.fn(),
}));

const mockedUseMarketStore = vi.mocked(useMarketStore);

describe("EvidenceRail", () => {
  it("renders compact evidence cards with deep links", () => {
    mockedUseMarketStore.mockReturnValue(
      buildMockMarketStoreState() as ReturnType<typeof useMarketStore>,
    );

    render(<EvidenceRail />);

    expect(screen.getByRole("heading", { name: "证据链" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "订单流" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "流动性" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "结构与微结构" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "查看信号深页" })).toHaveAttribute("href", "/signals");
    expect(screen.getByRole("link", { name: "查看市场深页" })).toHaveAttribute("href", "/market");
    expect(screen.getByRole("link", { name: "查看图表深页" })).toHaveAttribute("href", "/chart");
    expect(screen.getByText("净差")).toBeInTheDocument();
    expect(screen.getByText("盘口失衡")).toBeInTheDocument();
    expect(screen.getByText("趋势")).toBeInTheDocument();
  });
});
