import KlineChart from "@/components/chart/KlineChart";
import MarketSnapshotLoader from "@/components/market/MarketSnapshotLoader";
import PriceTicker from "@/components/market/PriceTicker";

export default function ChartPage() {
  return (
    <div className="space-y-6">
      <MarketSnapshotLoader />
      <PriceTicker />
      <KlineChart />
    </div>
  );
}
