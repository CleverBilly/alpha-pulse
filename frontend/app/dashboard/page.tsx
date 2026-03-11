import KlineChart from "@/components/chart/KlineChart";
import DecisionHeader from "@/components/dashboard/DecisionHeader";
import EvidenceRail from "@/components/dashboard/EvidenceRail";
import ExecutionPanel from "@/components/dashboard/ExecutionPanel";
import MarketSnapshotLoader from "@/components/market/MarketSnapshotLoader";

export default function DashboardPage() {
  return (
    <div className="space-y-6">
      <MarketSnapshotLoader />
      <DecisionHeader />
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-12">
        <div className="order-2 lg:order-1 lg:col-span-8">
          <KlineChart />
        </div>
        <div className="order-1 lg:order-2 lg:col-span-4">
          <ExecutionPanel />
        </div>
      </div>
      <EvidenceRail />
    </div>
  );
}
