import LiquidityPanel from "@/components/liquidity/LiquidityPanel";
import MarketLevelsBoard from "@/components/market/MarketLevelsBoard";
import MarketOverviewBoard from "@/components/market/MarketOverviewBoard";
import MarketSnapshotLoader from "@/components/market/MarketSnapshotLoader";
import MicrostructureTimeline from "@/components/market/MicrostructureTimeline";
import PriceTicker from "@/components/market/PriceTicker";
import SignalTape from "@/components/market/SignalTape";
import OrderFlowPanel from "@/components/orderflow/OrderFlowPanel";

export default function MarketPage() {
  return (
    <div className="space-y-6">
      <MarketSnapshotLoader />
      <PriceTicker />
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
