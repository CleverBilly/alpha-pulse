import FuturesWatchlist from "@/components/market/FuturesWatchlist";
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
      <FuturesWatchlist />
      <TradingWorkspaceHero
        eyebrow="Direction Drilldown"
        title="先看多周期方向，再下钻结构、流动性和订单流"
        description="市场页现在承接 Futures direction copilot。先用 watchlist 找到能跟的标的，再往下看结构层级、关键价位和微结构事件。"
        metrics={[
          { label: "Direction", value: "4h / 1h / 15m" },
          { label: "Trigger", value: "Flow + structure" },
          { label: "Liquidity", value: "Sweep + walls" },
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
