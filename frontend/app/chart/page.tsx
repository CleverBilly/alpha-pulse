import KlineChart from "@/components/chart/KlineChart";
import ChartInsightRail from "@/components/chart/ChartInsightRail";
import TradingWorkspaceHero from "@/components/layout/TradingWorkspaceHero";
import MarketSnapshotLoader from "@/components/market/MarketSnapshotLoader";

export default function ChartPage() {
  return (
    <div className="space-y-6">
      <MarketSnapshotLoader />
      <TradingWorkspaceHero
        eyebrow="图表"
        title="多图层图表工作区"
        description="专注图表本身，把结构、流动性和微结构图层叠加到同一张 K 线视图里。"
        metrics={[
          { label: "窗口", value: "48 根 K 线" },
          { label: "图层", value: "结构 + 流向" },
          { label: "模式", value: "手动下钻" },
        ]}
      />
      <div className="grid grid-cols-1 gap-6 xl:grid-cols-[minmax(0,1.45fr)_360px]">
        <KlineChart />
        <ChartInsightRail />
      </div>
    </div>
  );
}
