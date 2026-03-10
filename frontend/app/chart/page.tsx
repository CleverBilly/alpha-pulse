import KlineChart from "@/components/chart/KlineChart";
import ChartInsightRail from "@/components/chart/ChartInsightRail";
import TradingWorkspaceHero from "@/components/layout/TradingWorkspaceHero";
import MarketSnapshotLoader from "@/components/market/MarketSnapshotLoader";

export default function ChartPage() {
  return (
    <div className="space-y-6">
      <MarketSnapshotLoader />
      <TradingWorkspaceHero
        eyebrow="Chart"
        title="Multi-layer chart workspace"
        description="专注图表本身，把结构、流动性和微结构图层叠加到同一张 K 线视图里。"
        metrics={[
          { label: "Window", value: "48 candles" },
          { label: "Layers", value: "Structure + flow" },
          { label: "Mode", value: "Manual drill-down" },
        ]}
      />
      <div className="grid grid-cols-1 gap-6 xl:grid-cols-[minmax(0,1.45fr)_360px]">
        <KlineChart />
        <ChartInsightRail />
      </div>
    </div>
  );
}
