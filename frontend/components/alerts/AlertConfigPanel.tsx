"use client";

import type { ReactNode } from "react";
import { useEffect, useState } from "react";
import { Checkbox, Modal, Select, Slider, Spin, Switch, Tag } from "antd";
import type { AlertPreferences } from "@/types/alert";

type BrowserPermissionState = "unsupported" | "default" | "granted" | "denied";

interface AlertConfigPanelProps {
  open: boolean;
  loading: boolean;
  saving: boolean;
  preferences: AlertPreferences | null;
  browserPermission: BrowserPermissionState;
  onClose: () => void;
  onSave: (next: AlertPreferences) => Promise<void>;
}

export default function AlertConfigPanel({
  open,
  loading,
  saving,
  preferences,
  browserPermission,
  onClose,
  onSave,
}: AlertConfigPanelProps) {
  const [draft, setDraft] = useState<AlertPreferences | null>(preferences);

  useEffect(() => {
    if (open) {
      setDraft(preferences);
    }
  }, [open, preferences]);

  const current = draft ?? preferences;

  return (
    <Modal
      open={open}
      title="告警配置中心"
      okText="保存配置"
      cancelText="关闭"
      confirmLoading={saving}
      onCancel={onClose}
      onOk={async () => {
        if (!current) {
          return;
        }
        await onSave(current);
      }}
      className="app-shell__drawer"
    >
      {loading || !current ? (
        <div className="flex min-h-[220px] items-center justify-center">
          <Spin />
        </div>
      ) : (
        <div className="space-y-6">
          <section className="space-y-3 rounded-[24px] border border-slate-200 bg-slate-50/70 px-4 py-4">
            <div className="flex items-center justify-between gap-3">
              <div>
                <h3 className="text-sm font-semibold text-slate-950">推送渠道</h3>
                <p className="mt-1 text-xs leading-5 text-slate-500">飞书负责离站提醒，浏览器通知负责你正在打开页面时的即时提醒。</p>
              </div>
              <Tag color={resolvePermissionColor(browserPermission)}>{resolvePermissionLabel(browserPermission)}</Tag>
            </div>

            <Row
              label="飞书机器人"
              description="关闭后仍保留站内历史，但不再向飞书发送 webhook。"
              control={
                <Switch
                  checked={current.feishu_enabled}
                  onChange={(checked) => setDraft({ ...current, feishu_enabled: checked })}
                />
              }
            />
            <Row
              label="浏览器通知"
              description="关闭后告警中心仍保留 feed，但不会弹桌面提醒。"
              control={
                <Switch
                  checked={current.browser_enabled}
                  onChange={(checked) => setDraft({ ...current, browser_enabled: checked })}
                />
              }
            />
            <Row
              label="声音提示"
              description="新告警到达时播放提示音（需先点击页面激活音频上下文）。"
              control={
                <Switch
                  checked={current.sound_enabled ?? false}
                  onChange={(checked) => setDraft(current ? { ...current, sound_enabled: checked } : null)}
                  size="small"
                />
              }
            />
          </section>

          <section className="space-y-3 rounded-[24px] border border-slate-200 bg-white px-4 py-4">
            <div>
              <h3 className="text-sm font-semibold text-slate-950">事件类型</h3>
              <p className="mt-1 text-xs leading-5 text-slate-500">你不关心的事件直接在服务端过滤，减少站内 feed 和飞书噪音。</p>
            </div>
            <Checkbox
              checked={current.setup_ready_enabled}
              onChange={(event) => setDraft({ ...current, setup_ready_enabled: event.target.checked })}
            >
              A 级机会
            </Checkbox>
            <Checkbox
              checked={current.direction_shift_enabled}
              onChange={(event) => setDraft({ ...current, direction_shift_enabled: event.target.checked })}
            >
              方向切换
            </Checkbox>
            <Checkbox
              checked={current.no_trade_enabled}
              onChange={(event) => setDraft({ ...current, no_trade_enabled: event.target.checked })}
            >
              禁止交易
            </Checkbox>
          </section>

          <section className="space-y-4 rounded-[24px] border border-slate-200 bg-white px-4 py-4">
            <div>
              <h3 className="text-sm font-semibold text-slate-950">过滤条件</h3>
              <p className="mt-1 text-xs leading-5 text-slate-500">只保留你关心的标的和足够清晰的方向变化。</p>
            </div>

            <div>
              <p className="text-[11px] font-semibold uppercase tracking-[0.12em] text-slate-500">标的</p>
              <Checkbox.Group
                className="mt-3 flex flex-wrap gap-3"
                value={current.symbols}
                options={current.available_symbols.map((symbol) => ({
                  label: symbol,
                  value: symbol,
                }))}
                onChange={(next) => setDraft({ ...current, symbols: next.map(String) })}
              />
            </div>

            <div>
              <div className="flex items-center justify-between gap-3">
                <p className="text-[11px] font-semibold uppercase tracking-[0.12em] text-slate-500">最低置信度</p>
                <span className="text-sm font-semibold text-slate-950">{current.minimum_confidence}%</span>
              </div>
              <Slider
                min={0}
                max={100}
                step={1}
                value={current.minimum_confidence}
                onChange={(value) => setDraft({ ...current, minimum_confidence: Number(value) })}
              />
            </div>
          </section>

          <section className="space-y-4 rounded-[24px] border border-slate-200 bg-white px-4 py-4">
            <Row
              label="静默时段"
              description="静默时段只会跳过飞书推送，不会阻止站内历史落库。"
              control={
                <Switch
                  checked={current.quiet_hours_enabled}
                  onChange={(checked) => setDraft({ ...current, quiet_hours_enabled: checked })}
                />
              }
            />

            <div className="grid grid-cols-2 gap-3">
              <div>
                <p className="text-[11px] font-semibold uppercase tracking-[0.12em] text-slate-500">开始</p>
                <Select
                  className="mt-2 w-full"
                  value={current.quiet_hours_start}
                  options={buildHourOptions()}
                  onChange={(value) => setDraft({ ...current, quiet_hours_start: Number(value) })}
                  disabled={!current.quiet_hours_enabled}
                />
              </div>
              <div>
                <p className="text-[11px] font-semibold uppercase tracking-[0.12em] text-slate-500">结束</p>
                <Select
                  className="mt-2 w-full"
                  value={current.quiet_hours_end}
                  options={buildHourOptions()}
                  onChange={(value) => setDraft({ ...current, quiet_hours_end: Number(value) })}
                  disabled={!current.quiet_hours_enabled}
                />
              </div>
            </div>
          </section>
        </div>
      )}
    </Modal>
  );
}

function Row({
  label,
  description,
  control,
}: {
  label: string;
  description: string;
  control: ReactNode;
}) {
  return (
    <div className="flex items-start justify-between gap-4">
      <div className="min-w-0">
        <p className="text-sm font-semibold text-slate-950">{label}</p>
        <p className="mt-1 text-xs leading-5 text-slate-500">{description}</p>
      </div>
      <div className="shrink-0">{control}</div>
    </div>
  );
}

function buildHourOptions() {
  return Array.from({ length: 24 }).map((_, hour) => ({
    label: `${hour.toString().padStart(2, "0")}:00`,
    value: hour,
  }));
}

function resolvePermissionColor(permission: BrowserPermissionState) {
  if (permission === "granted") {
    return "success";
  }
  if (permission === "denied") {
    return "error";
  }
  if (permission === "default") {
    return "processing";
  }
  return "default";
}

function resolvePermissionLabel(permission: BrowserPermissionState) {
  if (permission === "granted") {
    return "浏览器权限已授权";
  }
  if (permission === "denied") {
    return "浏览器权限被拒绝";
  }
  if (permission === "default") {
    return "浏览器权限未授权";
  }
  return "浏览器通知不可用";
}
