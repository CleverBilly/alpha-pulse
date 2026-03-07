import AIAnalysisPanel from "@/components/analysis/AIAnalysisPanel";
import MarketSnapshotLoader from "@/components/market/MarketSnapshotLoader";
import PriceTicker from "@/components/market/PriceTicker";
import SignalCard from "@/components/signal/SignalCard";

export default function SignalsPage() {
  return (
    <div className="space-y-6">
      <MarketSnapshotLoader />
      <PriceTicker />
      <h2 className="text-2xl font-semibold">Trading Signals</h2>
      <div className="grid grid-cols-1 gap-6 xl:grid-cols-[0.9fr_1.1fr]">
        <SignalCard />
        <AIAnalysisPanel />
      </div>
    </div>
  );
}
