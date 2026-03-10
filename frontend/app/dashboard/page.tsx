import KlineChart from "@/components/chart/KlineChart";
import TradingWorkspaceHero from "@/components/layout/TradingWorkspaceHero";
import LiquidityPanel from "@/components/liquidity/LiquidityPanel";
import MarketSnapshotLoader from "@/components/market/MarketSnapshotLoader";
import OrderFlowPanel from "@/components/orderflow/OrderFlowPanel";
import SignalCard from "@/components/signal/SignalCard";

export default function DashboardPage() {
  return (
    <div className="space-y-6">
      <MarketSnapshotLoader />
      <TradingWorkspaceHero
        eyebrow="Dashboard"
        title="Spot market command deck"
        description="把价格、结构、订单流和信号压到同一块工作台里，适合快速确认当前 BTC / ETH 的执行偏向。"
        metrics={[
          { label: "Coverage", value: "BTC / ETH" },
          { label: "Core modules", value: "5 panels" },
          { label: "Feed mode", value: "Live snapshot" },
        ]}
      />
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-12">
        <div className="lg:col-span-8">
          <KlineChart />
        </div>
        <div className="lg:col-span-4">
          <SignalCard />
        </div>
      </div>
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-12">
        <div className="lg:col-span-5">
          <OrderFlowPanel />
        </div>
        <div className="lg:col-span-7">
          <LiquidityPanel />
        </div>
      </div>
    </div>
  );
}
