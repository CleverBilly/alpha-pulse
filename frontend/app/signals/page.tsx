import AIAnalysisPanel from "@/components/analysis/AIAnalysisPanel";
import TradingWorkspaceHero from "@/components/layout/TradingWorkspaceHero";
import MarketSnapshotLoader from "@/components/market/MarketSnapshotLoader";
import SignalCard from "@/components/signal/SignalCard";

export default function SignalsPage() {
  return (
    <div className="space-y-6">
      <MarketSnapshotLoader />
      <TradingWorkspaceHero
        eyebrow="Signals"
        title="Execution bias workspace"
        description="把模型输出、因子拆解和 AI 解释集中到一个更适合阅读的信号工作区。"
        metrics={[
          { label: "Signal card", value: "Primary" },
          { label: "AI memo", value: "Contextual" },
          { label: "Output", value: "Execution-ready" },
        ]}
      />
      <h2 className="text-2xl font-semibold">Trading Signals</h2>
      <div className="grid grid-cols-1 gap-6 xl:grid-cols-[0.9fr_1.1fr]">
        <SignalCard />
        <AIAnalysisPanel />
      </div>
    </div>
  );
}
