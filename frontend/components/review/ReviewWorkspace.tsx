import AIAnalysisPanel from "@/components/analysis/AIAnalysisPanel";
import AlertHistoryBoard from "@/components/alerts/AlertHistoryBoard";
import WinRatePanel from "@/components/alerts/WinRatePanel";
import CommandPage from "@/components/layout/CommandPage";
import OverviewBand from "@/components/layout/OverviewBand";
import TradingWorkspaceHero from "@/components/layout/TradingWorkspaceHero";
import MarketSnapshotLoader from "@/components/market/MarketSnapshotLoader";
import SignalCard from "@/components/signal/SignalCard";
import TradeOrderRail from "@/components/trading/TradeOrderRail";

const REVIEW_SYMBOLS = ["BTCUSDT", "ETHUSDT", "SOLUSDT"] as const;

export default function ReviewWorkspace() {
  return (
    <CommandPage className="review-command-page" data-testid="review-command-page">
      <MarketSnapshotLoader />
      <OverviewBand className="review-command-page__overview" data-testid="review-overview-band">
        <TradingWorkspaceHero
          eyebrow="复盘"
          title="信号复盘工作台"
          description="这里先看 A 级机会、禁止交易和方向切换的历史，再往下看当前实时信号与 AI 上下文，避免只盯最新一根 K 线。"
          metrics={[
            { label: "告警复盘", value: "持久化" },
            { label: "实时信号", value: "上下文化" },
            { label: "输出", value: "可复盘" },
          ]}
        />
      </OverviewBand>
      <div className="review-command-page__history" data-testid="review-history-strip">
        <WinRatePanel symbols={[...REVIEW_SYMBOLS]} />
        <AlertHistoryBoard />
      </div>
      <div className="review-command-page__context-head">
        <h2 className="text-2xl font-semibold">实时信号上下文</h2>
      </div>
      <div className="review-command-page__context" data-testid="review-context-workspace">
        <SignalCard />
        <AIAnalysisPanel />
      </div>
      <div className="review-command-page__trade-rail" data-testid="review-trade-rail">
        <TradeOrderRail />
      </div>
    </CommandPage>
  );
}
