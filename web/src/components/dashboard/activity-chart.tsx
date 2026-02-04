"use client";

import * as React from "react";
import { CartesianGrid, Line, LineChart, XAxis, YAxis } from "recharts";

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "~/components/ui/card";
import {
  ChartConfig,
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
} from "~/components/ui/chart";

const chartData = [
  { date: "2024-01-01", audits: 120, logins: 80 },
  { date: "2024-01-02", audits: 150, logins: 95 },
  { date: "2024-01-03", audits: 180, logins: 110 },
  { date: "2024-01-04", audits: 140, logins: 105 },
  { date: "2024-01-05", audits: 210, logins: 130 },
  { date: "2024-01-06", audits: 250, logins: 145 },
  { date: "2024-01-07", audits: 230, logins: 140 },
];

const chartConfig = {
  audits: {
    label: "Audits",
    color: "var(--color-audits)",
  },
  logins: {
    label: "Logins",
    color: "var(--color-logins)",
  },
} satisfies ChartConfig;

export function ActivityChart() {
  return (
    <Card className="flex flex-col">
      <CardHeader>
        <CardTitle>System Activity</CardTitle>
        <CardDescription>
          Daily trends for audit events and user logins.
        </CardDescription>
      </CardHeader>
      <CardContent className="flex-1 pb-0">
        <ChartContainer
          config={chartConfig}
          className="aspect-auto h-[250px] w-full"
        >
          <LineChart
            accessibilityLayer
            data={chartData}
            margin={{
              left: 12,
              right: 12,
            }}
          >
            <CartesianGrid vertical={false} />
            <XAxis
              dataKey="date"
              tickLine={false}
              axisLine={false}
              tickMargin={8}
              tickFormatter={(value) => {
                return new Date(value).toLocaleDateString("en-US", {
                  weekday: "short",
                });
              }}
            />
            <YAxis 
                tickLine={false}
                axisLine={false}
                tickMargin={8}
                fontSize={10}
            />
            <ChartTooltip
              cursor={false}
              content={<ChartTooltipContent hideLabel />}
            />
            <Line
              dataKey="audits"
              type="natural"
              stroke="var(--color-audits)"
              strokeWidth={2}
              dot={false}
            />
            <Line
              dataKey="logins"
              type="natural"
              stroke="var(--color-logins)"
              strokeWidth={2}
              dot={false}
            />
          </LineChart>
        </ChartContainer>
      </CardContent>
    </Card>
  );
}
