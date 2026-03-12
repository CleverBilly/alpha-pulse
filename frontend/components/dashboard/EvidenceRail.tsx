"use client";

import Link from "next/link";
import { useMarketStore } from "@/store/marketStore";
import { buildEvidenceSummary } from "./dashboardViewModel";

export default function EvidenceRail() {
  const { structure, liquidity, orderFlow, microstructureEvents } = useMarketStore();
  const cards = buildEvidenceSummary({
    structure,
    liquidity,
    orderFlow,
    microstructureEvents,
  });

  return (
    <section className="dashboard-evidence" aria-label="证据链">
      <div className="dashboard-evidence__header">
        <div>
          <p className="dashboard-evidence__eyebrow">证据链</p>
          <h2 className="dashboard-evidence__title">证据链</h2>
        </div>
        <p className="dashboard-evidence__description">只保留最影响短线判断的三组证据，并直接跳转到各自深页。</p>
      </div>

      <div className="dashboard-evidence__grid">
        {cards.map((card) => (
          <article key={card.id} className="dashboard-evidence__card surface-panel surface-panel--paper">
            <div className="dashboard-evidence__card-head">
              <div>
                <h3 className="dashboard-evidence__card-title">{card.title}</h3>
                <p className={`dashboard-evidence__state dashboard-evidence__state--${card.tone}`}>
                  {card.status === "ready" ? "已就绪" : "暂不可用"}
                </p>
              </div>
              <Link href={card.href} className="dashboard-evidence__link">
                {card.ctaLabel}
              </Link>
            </div>

            <p className="dashboard-evidence__summary">{card.summary}</p>

            <div className="dashboard-evidence__metrics">
              {card.metrics.length > 0 ? (
                card.metrics.map((metric) => (
                  <div key={`${card.id}-${metric.label}`} className="dashboard-evidence__metric">
                    <span>{metric.label}</span>
                    <strong>{metric.value}</strong>
                  </div>
                ))
              ) : (
                <div className="dashboard-evidence__metric dashboard-evidence__metric--empty">
                  <span>等待中</span>
                  <strong>--</strong>
                </div>
              )}
            </div>
          </article>
        ))}
      </div>
    </section>
  );
}
