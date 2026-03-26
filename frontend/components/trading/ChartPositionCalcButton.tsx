"use client";

import { useState } from "react";
import { Button, Drawer } from "antd";
import { CalculatorOutlined } from "@ant-design/icons";
import PositionCalculator from "@/components/trading/PositionCalculator";

export default function ChartPositionCalcButton() {
  const [calcOpen, setCalcOpen] = useState(false);

  return (
    <>
      <div style={{ position: "fixed", right: 24, bottom: 80, zIndex: 100 }}>
        <Button
          type="primary"
          shape="circle"
          icon={<CalculatorOutlined />}
          size="large"
          onClick={() => setCalcOpen(true)}
          aria-label="打开仓位计算器"
        />
      </div>
      <Drawer
        title="仓位计算器"
        open={calcOpen}
        onClose={() => setCalcOpen(false)}
        width={340}
        placement="right"
      >
        <PositionCalculator />
      </Drawer>
    </>
  );
}
