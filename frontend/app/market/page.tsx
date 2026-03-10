import LiquidityPanel from "@/components/liquidity/LiquidityPanel";
import TradingWorkspaceHero from "@/components/layout/TradingWorkspaceHero";
import MarketLevelsBoard from "@/components/market/MarketLevelsBoard";
import MarketOverviewBoard from "@/components/market/MarketOverviewBoard";
import MarketSnapshotLoader from "@/components/market/MarketSnapshotLoader";
import MicrostructureTimeline from "@/components/market/MicrostructureTimeline";
import SignalTape from "@/components/market/SignalTape";
import OrderFlowPanel from "@/components/orderflow/OrderFlowPanel";

export default function MarketPage() {
  return (
    <div className="space-y-6">
      <MarketSnapshotLoader />
      <TradingWorkspaceHero
        eyebrow="Market"
        title="Structure, ladder and flow in one view"
        description="这里偏向盘中判断。你可以在同一个页面里看总览、关键价位、信号带和微结构时间线。"
        metrics={[
          { label: "Overview", value: "Regime + levels" },
          { label: "Timeline", value: "Micro events" },
          { label: "Liquidity", value: "Wall map" },
        ]}
      />
      <MarketOverviewBoard />
      <div className="grid grid-cols-1 gap-6 xl:grid-cols-[1.15fr_0.85fr]">
        <MarketLevelsBoard />
        <div className="space-y-6">
          <SignalTape />
          <MicrostructureTimeline />
        </div>
      </div>
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        <OrderFlowPanel />
        <LiquidityPanel />
      </div>
    </div>
  );
}
