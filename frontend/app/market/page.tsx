import FuturesWatchlist from "@/components/market/FuturesWatchlist";
import LiquidityPanel from "@/components/liquidity/LiquidityPanel";
import CommandPage from "@/components/layout/CommandPage";
import OverviewBand from "@/components/layout/OverviewBand";
import TradingWorkspaceHero from "@/components/layout/TradingWorkspaceHero";
import MarketLevelsBoard from "@/components/market/MarketLevelsBoard";
import MarketOverviewBoard from "@/components/market/MarketOverviewBoard";
import MarketSnapshotLoader from "@/components/market/MarketSnapshotLoader";
import MicrostructureTimeline from "@/components/market/MicrostructureTimeline";
import SignalTape from "@/components/market/SignalTape";
import OrderFlowPanel from "@/components/orderflow/OrderFlowPanel";

export default function MarketPage() {
  return (
    <CommandPage className="market-command-page" data-testid="market-command-page">
      <MarketSnapshotLoader />
      <div className="market-command-page__watchlist" data-testid="market-watchlist-band">
        <FuturesWatchlist />
      </div>
      <OverviewBand className="market-command-page__overview" data-testid="market-overview-band">
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
      </OverviewBand>
      <div className="market-command-page__primary-grid" data-testid="market-primary-grid">
        <div className="market-command-page__levels-stage">
          <MarketLevelsBoard />
        </div>
        <div className="market-command-page__secondary-rail" data-testid="market-secondary-rail">
          <SignalTape />
          <MicrostructureTimeline />
        </div>
      </div>
      <div className="market-command-page__diagnostics" data-testid="market-diagnostics-strip">
        <OrderFlowPanel />
        <LiquidityPanel />
      </div>
    </CommandPage>
  );
}
