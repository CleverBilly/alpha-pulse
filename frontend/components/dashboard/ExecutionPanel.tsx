"use client";

import { useMarketStore } from "@/store/marketStore";
import { buildDashboardDecision, buildDirectionAwareExecutionSetup, buildDirectionCopilotDecision } from "./dashboardViewModel";

export default function ExecutionPanel() {
  const {
    signal,
    structure,
    liquidity,
    orderFlow,
    microstructureEvents,
    price,
    directionSnapshots,
  } = useMarketStore();

  const decision =
    directionSnapshots.macro && directionSnapshots.bias && directionSnapshots.trigger && directionSnapshots.execution
      ? buildDirectionCopilotDecision({
          macroSnapshot: directionSnapshots.macro,
          biasSnapshot: directionSnapshots.bias,
          triggerSnapshot: directionSnapshots.trigger,
          executionSnapshot: directionSnapshots.execution,
        })
      : buildDashboardDecision({
          signal,
          structure,
          liquidity,
          orderFlow,
        });

  const setup = buildDirectionAwareExecutionSetup({
    signal,
    structure,
    liquidity,
    orderFlow,
    microstructureEvents,
    price,
    decision,
  });

  return (
    <section className="dashboard-setup surface-panel surface-panel--paper">
      <div className="dashboard-setup__header">
        <div>
          <p className="dashboard-setup__eyebrow">执行线索</p>
          <h2 className="dashboard-setup__title">执行方案</h2>
        </div>
        <span className={`dashboard-setup__badge dashboard-setup__badge--${setup.tone}`}>{setup.biasLabel}</span>
      </div>

      {setup.status === "ready" ? (
        <div className="dashboard-setup__body">
          <p className="dashboard-setup__description">{setup.reason}</p>

          <div className="dashboard-setup__grid">
            <MetricCard label="进场区间" value={`${setup.entryLow.toFixed(2)} - ${setup.entryHigh.toFixed(2)}`} />
            <MetricCard label="止损位" value={setup.stopLoss.toFixed(2)} />
            <MetricCard label="目标位" value={setup.target.toFixed(2)} />
            <MetricCard label="R / R" value={setup.riskReward.toFixed(2)} accent />
          </div>

          <div className="dashboard-setup__block">
            <p className="dashboard-setup__block-label">触发条件</p>
            <p className="dashboard-setup__block-copy">{setup.trigger}</p>
          </div>

          <div className="dashboard-setup__block dashboard-setup__block--warning">
            <p className="dashboard-setup__block-label">风险提示</p>
            <p className="dashboard-setup__block-copy">{setup.caution}</p>
          </div>
        </div>
      ) : (
        <div className="dashboard-setup__empty">
          <p className="dashboard-setup__description">{setup.reason}</p>
          <p className="dashboard-setup__block-copy">{setup.trigger}</p>
          <button type="button" disabled className="dashboard-setup__disabled">
            等待 setup 完整
          </button>
        </div>
      )}
    </section>
  );
}

function MetricCard({
  label,
  value,
  accent = false,
}: {
  label: string;
  value: string;
  accent?: boolean;
}) {
  return (
    <div className={`dashboard-setup__metric ${accent ? "dashboard-setup__metric--accent" : ""}`}>
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}
