import KlineChart from "@/components/chart/KlineChart";
import ChartInsightRail from "@/components/chart/ChartInsightRail";
import CommandPage from "@/components/layout/CommandPage";
import OverviewBand from "@/components/layout/OverviewBand";
import TradingWorkspaceHero from "@/components/layout/TradingWorkspaceHero";
import MarketSnapshotLoader from "@/components/market/MarketSnapshotLoader";
import ChartPositionCalcButton from "@/components/trading/ChartPositionCalcButton";

export default function ChartPage() {
  return (
    <CommandPage className="chart-command-page" data-testid="chart-command-page">
      <MarketSnapshotLoader />
      <OverviewBand className="chart-command-page__overview" data-testid="chart-overview-band">
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
      </OverviewBand>
      <div className="chart-command-page__workspace" data-testid="chart-primary-workspace">
        <div className="chart-command-page__chart-stage">
          <KlineChart />
        </div>
        <div className="chart-command-page__insight-rail" data-testid="chart-insight-side-rail">
          <ChartInsightRail />
        </div>
      </div>
      <div className="chart-command-page__actions" data-testid="chart-action-strip">
        <ChartPositionCalcButton />
      </div>
    </CommandPage>
  );
}
