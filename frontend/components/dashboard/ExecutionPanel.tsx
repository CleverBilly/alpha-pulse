"use client";

import { useMarketStore } from "@/store/marketStore";
import { buildExecutionSetup } from "./dashboardViewModel";

export default function ExecutionPanel() {
  const { signal, structure, liquidity, orderFlow, microstructureEvents, price } = useMarketStore();
  const setup = buildExecutionSetup({
    signal,
    structure,
    liquidity,
    orderFlow,
    microstructureEvents,
    price,
  });

  return (
    <section className="dashboard-setup surface-panel surface-panel--paper">
      <div className="dashboard-setup__header">
        <div>
          <p className="dashboard-setup__eyebrow">执行线索</p>
          <h2 className="dashboard-setup__title">Execution Setup</h2>
        </div>
        <span className={`dashboard-setup__badge dashboard-setup__badge--${setup.tone}`}>{setup.biasLabel}</span>
      </div>

      {setup.status === "ready" ? (
        <div className="dashboard-setup__body">
          <p className="dashboard-setup__description">{setup.reason}</p>

          <div className="dashboard-setup__grid">
            <MetricCard label="Entry Zone" value={`${setup.entryLow.toFixed(2)} - ${setup.entryHigh.toFixed(2)}`} />
            <MetricCard label="Stop Loss" value={setup.stopLoss.toFixed(2)} />
            <MetricCard label="Target" value={setup.target.toFixed(2)} />
            <MetricCard label="R / R" value={setup.riskReward.toFixed(2)} accent />
          </div>

          <div className="dashboard-setup__block">
            <p className="dashboard-setup__block-label">Trigger</p>
            <p className="dashboard-setup__block-copy">{setup.trigger}</p>
          </div>

          <div className="dashboard-setup__block dashboard-setup__block--warning">
            <p className="dashboard-setup__block-label">Risk Note</p>
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
