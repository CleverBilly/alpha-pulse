import AlertCenter from "@/components/alerts/AlertCenter";
import AlertHistoryBoard from "@/components/alerts/AlertHistoryBoard";
import CommandPage from "@/components/layout/CommandPage";
import OverviewBand from "@/components/layout/OverviewBand";
import CommandPanel from "@/components/layout/CommandPanel";

const ALERT_METRICS = [
  { label: "通道", value: "站内 / 浏览器 / 飞书" },
  { label: "实时策略", value: "A 级机会、方向切换、禁止交易" },
  { label: "工作方式", value: "实时值班 + 历史复盘" },
];

export default function AlertsPage() {
  return (
    <CommandPage className="alerts-command-page" data-testid="alerts-command-page">
      <OverviewBand className="alerts-command-page__overview" data-testid="alerts-overview-band">
        <CommandPanel className="alerts-status-band" data-testid="alerts-status-band" data-surface="console" variant="control">
          <div className="alerts-status-band__copy">
            <p className="alerts-status-band__eyebrow">告警中枢</p>
            <h1 className="alerts-status-band__title">值班控制、实时事件流与复盘轨道</h1>
            <p className="alerts-status-band__description">
              把提醒系统收成真正的调度中枢。左侧先处理现在要不要行动，右侧再回看事件为什么发生。
            </p>
          </div>

          <div className="alerts-status-band__metrics">
            {ALERT_METRICS.map((metric) => (
              <div key={metric.label} className="alerts-status-band__metric">
                <span>{metric.label}</span>
                <strong>{metric.value}</strong>
              </div>
            ))}
          </div>
        </CommandPanel>
      </OverviewBand>

      <div className="alerts-command-page__workspace">
        <CommandPanel
          className="alerts-command-page__watch-desk"
          data-testid="alerts-watch-desk"
          variant="control"
          data-surface="console"
        >
          <div className="alerts-command-page__control-copy">
            <p className="alerts-command-page__eyebrow">快速控制</p>
            <h2 className="alerts-command-page__title">实时值班台</h2>
            <p className="alerts-command-page__description">
              保留全局抽屉的快捷入口，但把页面本身升级成完整工作台，先做即时处理，再衔接历史复盘。
            </p>
          </div>
          <AlertCenter mode="page" />
        </CommandPanel>

        <div className="alerts-command-page__history-rail" data-testid="alerts-history-rail">
          <AlertHistoryBoard />
        </div>
      </div>
    </CommandPage>
  );
}
