/**
 * 根据 alert 的 verdict 或 tradeability_label 判断是否为多头方向。
 * Task 5 (KlineChart) 和 Task 11 (ReviewChartModal) 共享此工具函数。
 */
export function isLongDirection(verdict?: string, tradeabilityLabel?: string): boolean {
  const text = `${verdict ?? ""} ${tradeabilityLabel ?? ""}`.toLowerCase();
  return text.includes("long") || text.includes("多") || text.includes("bullish");
}
