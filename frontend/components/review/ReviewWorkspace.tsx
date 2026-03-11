import AIAnalysisPanel from "@/components/analysis/AIAnalysisPanel";
import AlertHistoryBoard from "@/components/alerts/AlertHistoryBoard";
import TradingWorkspaceHero from "@/components/layout/TradingWorkspaceHero";
import MarketSnapshotLoader from "@/components/market/MarketSnapshotLoader";
import SignalCard from "@/components/signal/SignalCard";

export default function ReviewWorkspace() {
  return (
    <div className="space-y-6">
      <MarketSnapshotLoader />
      <TradingWorkspaceHero
        eyebrow="Review"
        title="Signal review workspace"
        description="这里先看 A 级 setup、No-Trade 和方向切换的历史，再往下看当前 live signal 与 AI context，避免只盯最新一根 K 线。"
        metrics={[
          { label: "Alert replay", value: "Persistent" },
          { label: "Live signal", value: "Contextual" },
          { label: "Output", value: "Review-ready" },
        ]}
      />
      <AlertHistoryBoard />
      <h2 className="text-2xl font-semibold">Live Signal Context</h2>
      <div className="grid grid-cols-1 gap-6 xl:grid-cols-[0.9fr_1.1fr]">
        <SignalCard />
        <AIAnalysisPanel />
      </div>
    </div>
  );
}
