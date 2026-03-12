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
        eyebrow="方向下钻"
        title="先看多周期方向，再下钻结构、流动性和订单流"
        description="市场页现在承接合约方向副驾驶。先用观察列表找到能跟的标的，再往下看结构层级、关键价位和微结构事件。"
        metrics={[
          { label: "方向", value: "4h / 1h / 15m / 5m" },
          { label: "触发", value: "15m + 5m 执行" },
          { label: "流动性", value: "扫流动性 + 墙位" },
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
