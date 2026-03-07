import KlineChart from "@/components/chart/KlineChart";
import LiquidityPanel from "@/components/liquidity/LiquidityPanel";
import MarketSnapshotLoader from "@/components/market/MarketSnapshotLoader";
import PriceTicker from "@/components/market/PriceTicker";
import OrderFlowPanel from "@/components/orderflow/OrderFlowPanel";
import SignalCard from "@/components/signal/SignalCard";

export default function DashboardPage() {
  return (
    <div className="space-y-6">
      <MarketSnapshotLoader />
      <PriceTicker />
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
        <div className="lg:col-span-2">
          <KlineChart />
        </div>
        <SignalCard />
      </div>
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        <OrderFlowPanel />
        <LiquidityPanel />
      </div>
    </div>
  );
}
