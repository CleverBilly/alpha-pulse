"use client";

import { useEffect, useState } from "react";
import { Button, Spin, Statistic, Table, Tag } from "antd";
import { alertApi } from "@/services/apiClient";
import type { AlertStats } from "@/types/alert";

const LIMIT_OPTIONS = [20, 50, 0] as const; // 0 = 全部
type Limit = (typeof LIMIT_OPTIONS)[number];

interface WinRatePanelProps {
  symbols: string[];
}

export default function WinRatePanel({ symbols }: WinRatePanelProps) {
  const [limit, setLimit] = useState<Limit>(50);
  const [stats, setStats] = useState<AlertStats[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let active = true;
    setLoading(true);
    Promise.all(symbols.map((s) => alertApi.getAlertStats(s, limit)))
      .then((results) => {
        if (active) setStats(results);
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, [symbols, limit]);

  const columns = [
    {
      title: "标的",
      dataIndex: "symbol",
      key: "symbol",
      render: (v: string) => <Tag>{v.replace("USDT", "")}</Tag>,
    },
    {
      title: "胜率",
      dataIndex: "win_rate",
      key: "win_rate",
      render: (v: number) => (
        <Statistic
          value={v}
          precision={1}
          suffix="%"
          valueStyle={{ fontSize: 14, color: v >= 55 ? "#52c41a" : "#ff7875" }}
        />
      ),
    },
    {
      title: "平均 R:R",
      dataIndex: "avg_rr",
      key: "avg_rr",
      render: (v: number) => (v > 0 ? v.toFixed(2) : "—"),
    },
    {
      title: "命中/止损",
      key: "hits",
      render: (_: unknown, row: AlertStats) => `${row.target_hit} / ${row.stop_hit}`,
    },
    {
      title: "样本",
      dataIndex: "sample_size_label",
      key: "sample_size_label",
    },
  ];

  return (
    <div style={{ marginBottom: 24 }}>
      <div style={{ display: "flex", alignItems: "center", gap: 8, marginBottom: 12 }}>
        <span style={{ fontWeight: 600 }}>历史胜率统计</span>
        {LIMIT_OPTIONS.map((l) => (
          <Button
            key={l}
            size="small"
            type={limit === l ? "primary" : "default"}
            onClick={() => setLimit(l)}
          >
            {l === 0 ? "全部" : `近${l}条`}
          </Button>
        ))}
      </div>
      {loading ? (
        <Spin />
      ) : (
        <Table
          dataSource={stats}
          columns={columns}
          rowKey="symbol"
          pagination={false}
          size="small"
        />
      )}
    </div>
  );
}
