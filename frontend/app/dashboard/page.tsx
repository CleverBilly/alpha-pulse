import KlineChart from "@/components/chart/KlineChart";
import DecisionHeader from "@/components/dashboard/DecisionHeader";
import EvidenceRail from "@/components/dashboard/EvidenceRail";
import ExecutionPanel from "@/components/dashboard/ExecutionPanel";
import CommandPage from "@/components/layout/CommandPage";
import OverviewBand from "@/components/layout/OverviewBand";
import MarketSnapshotLoader from "@/components/market/MarketSnapshotLoader";
import PositionCalculator from "@/components/trading/PositionCalculator";

export default function DashboardPage() {
  return (
    <CommandPage className="dashboard-command-page" data-testid="dashboard-command-page">
      <MarketSnapshotLoader />
      <OverviewBand className="dashboard-command-page__overview" data-testid="dashboard-overview-band">
        <DecisionHeader />
      </OverviewBand>
      <div className="dashboard-command-page__workspace" data-testid="dashboard-primary-workspace">
        <div className="dashboard-command-page__chart-stage" data-testid="dashboard-chart-surface">
          <KlineChart />
        </div>
        <div className="dashboard-command-page__side-rail" data-testid="dashboard-side-rail">
          <ExecutionPanel />
          <PositionCalculator />
        </div>
      </div>
      <div className="dashboard-command-page__evidence" data-testid="dashboard-evidence-surface">
        <EvidenceRail />
      </div>
    </CommandPage>
  );
}
