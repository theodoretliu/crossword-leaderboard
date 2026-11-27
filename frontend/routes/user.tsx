import { getRouteApi } from "@tanstack/react-router";
import dayjs from "dayjs";
import utc from "dayjs/plugin/utc";
import { Header } from "@/components/header";
import { API_URL } from "@/api";
import * as z from "zod";
import useSWR from "swr";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { dateToFormat, secondsToMinutes } from "@/utils";

import {
  BarChart,
  Bar,
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Scatter,
  RadarChart,
  Radar,
  PolarGrid,
  PolarAngleAxis,
  PolarRadiusAxis,
} from "recharts";
import {
  ChartConfig,
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
} from "@/components/ui/chart";
import { CardTitle } from "@/components/ui/card";

const chartConfig = {
  t: {
    label: "Time in seconds",
    color: "hsl(var(--chart-1))",
  },
  average100: {
    label: "100-day moving average",
    color: "hsl(var(--chart-2))",
  },
  average200: {
    label: "200-day moving average",
    color: "hsl(var(--chart-3))",
  },
} satisfies ChartConfig;

const radarConfig = {
  percentile: {
    label: "You",
    color: "hsl(var(--chart-1))",
  },
  median: {
    label: "50th percentile",
    color: "hsl(var(--muted-foreground))",
  },
} satisfies ChartConfig;

// ISO weekday order, matching the API (1 = Monday).
const WEEKDAYS = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"];

type RadarDatum = {
  day: string;
  // Share of everyone else's solves on this weekday slower than this user's
  // median, 0-100.
  percentile: number | null;
  median: number;
  medianTime: number | null;
  solves: number;
};

function RadarTooltip({
  active,
  payload,
}: {
  active?: boolean;
  payload?: { payload: RadarDatum }[];
}) {
  if (!active || !payload?.length) {
    return null;
  }
  const { day, percentile, medianTime, solves } = payload[0].payload;
  if (percentile === null || medianTime === null) {
    return (
      <div className="rounded-lg border bg-background px-2.5 py-1.5 text-xs shadow-xl">
        <div className="font-medium">{day}</div>
        <div>No solves yet</div>
      </div>
    );
  }

  return (
    <div className="rounded-lg border bg-background px-2.5 py-1.5 text-xs shadow-xl">
      <div className="font-medium">{day}</div>
      <div>Median {secondsToMinutes(medianTime)}</div>
      <div>Faster than {Math.round(percentile)}% of everyone else&apos;s solves</div>
      <div className="text-muted-foreground">
        {solves} {solves === 1 ? "solve" : "solves"}
      </div>
    </div>
  );
}

const StatsType = z.object({
  Average: z.number(),
  Median: z.number(),
  Best: z.number(),
  Worst: z.number(),
  NumCompleted: z.number(),
});

const ResponseType = z.object({
  Username: z.string(),
  MiniStats: StatsType,
  SaturdayStats: StatsType,
  OverallStats: StatsType,
  LongestStreak: z.number(),
  CurrentStreak: z.number(),
  AllTimes: z.array(
    z.object({
      t: z.number(),
      d: z.coerce.date(),
    }),
  ),
  DayOfWeek: z.array(
    z.object({
      Weekday: z.number(),
      MedianTime: z.number(),
      NumSolves: z.number(),
      Percentile: z.number(),
    }),
  ),
});

dayjs.extend(utc);

async function fetcher(key: string) {
  const res = await fetch(API_URL + key);
  const json = await res.json();

  return ResponseType.parse(json);
}

const route = getRouteApi("/users/$userId");

export function User() {
  const { userId } = route.useParams();

  const { error, data } = useSWR(`/users/${userId}`, fetcher);

  const allTimes = data?.AllTimes ?? [];

  const movingAverage = allTimes.map((solve, index) => {
    const window100 = allTimes.slice(Math.max(0, index - 100), index + 1);
    const window200 = allTimes.slice(Math.max(0, index - 200), index + 1);
    const average100 =
      window100.reduce((acc, curr) => {
        return acc + (dayjs(curr.d).utc().day() === 6 ? curr.t * (25 / 49) : curr.t);
      }, 0) / window100.length;

    const average200 =
      window200.reduce((acc, curr) => {
        return acc + (dayjs(curr.d).utc().day() === 6 ? curr.t * (25 / 49) : curr.t);
      }, 0) / window200.length;

    return {
      d: solve.d,
      average100,
      average200,
      t: solve.t,
    };
  });

  const radarData: RadarDatum[] = WEEKDAYS.map((day, i) => {
    const row = data?.DayOfWeek.find((r) => r.Weekday === i + 1);
    return {
      day,
      percentile: row?.Percentile ?? null,
      median: 50,
      medianTime: row?.MedianTime ?? null,
      solves: row?.NumSolves ?? 0,
    };
  });

  if (error) {
    return "Error";
  }

  if (!data) {
    return <div />;
  }

  return (
    <div>
      <Header />

      <div className="flex max-w-[1024px] flex-col gap-4">
        <h2 className="px-4 text-lg font-semibold">Statistics for {data.Username}</h2>

        <div className="flex flex-row items-center gap-4 px-4">
          <h3>Current Streak: {data.CurrentStreak}</h3>

          <h3>Longest Streak: {data.LongestStreak}</h3>
        </div>

        <h2 className="px-4 text-lg font-semibold">Cumulative Statistics</h2>

        <div className="w-full overflow-auto">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Category</TableHead>

                <TableHead>Median Time (s)</TableHead>
                <TableHead>Best Time (s)</TableHead>
                <TableHead>Worst Time (s)</TableHead>
                <TableHead>Number completed</TableHead>
              </TableRow>
            </TableHeader>

            <TableBody>
              <TableRow>
                <TableCell>Minis (5x5)</TableCell>
                <TableCell>{secondsToMinutes(data.MiniStats.Median, true)}</TableCell>
                <TableCell>{secondsToMinutes(data.MiniStats.Best)}</TableCell>
                <TableCell>{secondsToMinutes(data.MiniStats.Worst)}</TableCell>
                <TableCell>{data.MiniStats.NumCompleted}</TableCell>
              </TableRow>

              <TableRow>
                <TableCell>Saturdays (7x7)</TableCell>
                <TableCell>{secondsToMinutes(data.SaturdayStats.Median, true)}</TableCell>
                <TableCell>{secondsToMinutes(data.SaturdayStats.Best)}</TableCell>
                <TableCell>{secondsToMinutes(data.SaturdayStats.Worst)}</TableCell>
                <TableCell>{data.SaturdayStats.NumCompleted}</TableCell>
              </TableRow>

              <TableRow>
                <TableCell>Overall</TableCell>
                <TableCell>{secondsToMinutes(data.OverallStats.Median, true)}</TableCell>
                <TableCell>{secondsToMinutes(data.OverallStats.Best)}</TableCell>
                <TableCell>{secondsToMinutes(data.OverallStats.Worst)}</TableCell>
                <TableCell>{data.OverallStats.NumCompleted}</TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </div>

        {data.DayOfWeek.length > 0 && (
          <div className="p-4 pt-0">
            <h2 className="text-lg font-semibold">Speed by Day of Week</h2>
            <p className="mb-2 text-sm text-muted-foreground">
              Your median time on each day of the week, as a percentile of everyone else&apos;s
              solves on that day. Further out is faster; the dashed ring is the 50th percentile.
            </p>

            <ChartContainer config={radarConfig} className="mx-auto aspect-square max-h-[360px]">
              <RadarChart data={radarData} outerRadius="75%">
                <PolarGrid />
                <PolarAngleAxis dataKey="day" />
                <PolarRadiusAxis
                  domain={[0, 100]}
                  tickCount={5}
                  angle={90 - 180 / WEEKDAYS.length}
                  tick={{ fontSize: 10 }}
                  axisLine={false}
                />
                <ChartTooltip content={<RadarTooltip />} />
                <Radar
                  dataKey="median"
                  stroke="var(--color-median)"
                  strokeDasharray="4 4"
                  fill="none"
                  isAnimationActive={false}
                />
                <Radar
                  dataKey="percentile"
                  stroke="var(--color-percentile)"
                  fill="var(--color-percentile)"
                  fillOpacity={0.3}
                  strokeWidth={2}
                  dot={{ r: 3 }}
                />
              </RadarChart>
            </ChartContainer>
          </div>
        )}

        <div className="p-4 pt-0">
          <h2 className="mb-4 text-lg font-semibold">Solve Time History</h2>

          <ChartContainer config={chartConfig} className="min-h-[300px] w-full">
            <LineChart data={movingAverage}>
              <XAxis
                dataKey="d"
                scale="time"
                interval="preserveStartEnd"
                tickFormatter={(value) => dayjs(value).utc().format("M/D/YY")}
              />

              <YAxis
                label={{
                  value: "Time (seconds)",
                  angle: -90,
                  position: "insideLeft",
                }}
                interval="preserveStart"
                domain={[0, Math.max(...movingAverage.map((solve) => solve.average100)) * 1.1]}
                allowDataOverflow
              />

              <ChartTooltip
                content={
                  <ChartTooltipContent
                    // labelKey="d"
                    labelFormatter={(value) => dayjs(value).utc().format("MMM D, YYYY")}
                  />
                }
              />

              <Line
                dataKey="t"
                stroke="var(--color-t)"
                strokeOpacity={0.2}
                strokeWidth={0}
                dot={{ stroke: "var(--color-t)", strokeWidth: 2 }}
                type="monotone"
              />

              <Line
                dataKey="average100"
                stroke="var(--color-average100)"
                strokeWidth={2}
                strokeOpacity={0.5}
                dot={false}
                type="monotone"
              />

              <Line
                dataKey="average200"
                stroke="var(--color-average200)"
                strokeWidth={2}
                dot={false}
                type="monotone"
              />

              <Tooltip />

              <CartesianGrid strokeDasharray="3 3" />
            </LineChart>
          </ChartContainer>
        </div>
      </div>
    </div>
  );
}
