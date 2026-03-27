import CommandPage from "@/components/layout/CommandPage";
import OverviewBand from "@/components/layout/OverviewBand";
import CommandPanel from "@/components/layout/CommandPanel";

const AUTO_TRADING_PANELS = [
  {
    title: "策略编排",
    description: "这里会承接策略列表、启停开关、执行时段与策略级别优先级。",
  },
  {
    title: "风险护栏",
    description: "统一收口最大风险、单日停机条件、异常熔断与手动接管入口。",
  },
  {
    title: "运行状态",
    description: "后续展示当前执行中的策略、健康检查和最近一次同步结果。",
  },
  {
    title: "动作记录",
    description: "预留最近执行动作、风险触发记录和人工干预日志区域。",
  },
];

export default function AutoTradingPage() {
  return (
    <CommandPage className="auto-trading-command-page" data-testid="auto-trading-command-page">
      <OverviewBand className="auto-trading-command-page__overview" data-testid="auto-trading-overview-band">
        <div className="page-hero">
          <div>
            <p className="page-hero__eyebrow">执行中枢</p>
            <h1 className="page-hero__title">自动交易指挥台</h1>
            <p className="page-hero__description">
              自动交易功能还在建设中，但这页先定义它在整套系统里的位置：策略编排、运行状态、风险护栏和动作记录都将在这里汇合。
            </p>
          </div>
          <div className="page-hero__metrics">
            <div className="page-hero__metric">
              <span>状态</span>
              <strong>建设中</strong>
            </div>
            <div className="page-hero__metric">
              <span>核心能力</span>
              <strong>策略 / 风控 / 审计</strong>
            </div>
            <div className="page-hero__metric">
              <span>当前目标</span>
              <strong>先定母版页</strong>
            </div>
          </div>
        </div>
      </OverviewBand>

      <div className="auto-trading-command-page__workspace" data-testid="auto-trading-workspace">
        {AUTO_TRADING_PANELS.map((panel) => (
          <CommandPanel key={panel.title} className="auto-trading-command-page__panel" variant="control">
            <p className="alerts-command-page__eyebrow">预留模块</p>
            <h2 className="alerts-command-page__title">{panel.title}</h2>
            <p className="alerts-command-page__description">{panel.description}</p>
          </CommandPanel>
        ))}
      </div>
    </CommandPage>
  );
}
