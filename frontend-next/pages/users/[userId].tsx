/** @jsx jsx */
import { jsx } from "@emotion/react";
import { useRouter } from "next/router";
import dayjs from "dayjs";
import utc from "dayjs/plugin/utc";
import { GetServerSideProps } from "next";
import { Header } from "components/header";
import { API_URL } from "api";
import * as s from "superstruct";
import { H2 } from "components/h2";
import useSWR from "swr";
import {
  Paper,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
} from "@material-ui/core";
import { dateToFormat, secondsToMinutes } from "utils";
import * as styles from "components/[userId]_styles";

import {
  ResponsiveContainer,
  LineChart,
  XAxis,
  YAxis,
  Tooltip,
  Line,
  Label,
} from "recharts";
import { useFeatureFlag } from "hooks/use_feature_flag";

const StatsType = s.object({
  Average: s.number(),
  Best: s.number(),
  Worst: s.number(),
  NumCompleted: s.number(),
});

const ResponseType = s.object({
  Username: s.string(),
  MiniStats: StatsType,
  SaturdayStats: StatsType,
  OverallStats: StatsType,
  LongestStreak: s.number(),
  CurrentStreak: s.number(),
  EloHistory: s.array(
    s.object({
      Date: s.string(),
      Elo: s.number(),
    })
  ),
  PeakElo: s.number(),
  CurrentElo: s.number(),
});

dayjs.extend(utc);

async function fetcher(key: string) {
  const res = await fetch(API_URL + key);
  const json = await res.json();

  s.assert(json, ResponseType);

  return json;
}

export const getServerSideProps: GetServerSideProps = async (context) => {
  const { userId } = context.query;
  const initialData = await fetcher(`/users/${userId}`);

  return { props: { initialData } };
};

export default function User({
  initialData,
}: {
  initialData: s.Infer<typeof ResponseType>;
}) {
  const router = useRouter();

  const { userId } = router.query;

  const { error, data } = useSWR(`/users/${userId}`, fetcher, { initialData });

  const {
    error: eloFFError,
    loading: eloFFLoading,
    status: eloFF,
  } = useFeatureFlag("elos");

  if (error || eloFFError) {
    return "Error";
  }

  if (!data || eloFFLoading) {
    return <div />;
  }

  return (
    <div>
      <Header />
      <H2>Statistics for {data.Username}</H2>

      <div css={styles.streak}>
        <h3>Current Streak: {data.CurrentStreak}</h3>

        <h3>Longest Streak: {data.LongestStreak}</h3>
      </div>

      {eloFF && (
        <div css={styles.streak}>
          <h3>Current ELO: {data.CurrentElo.toFixed(2)}</h3>

          <h3>Peak ELO: {data.PeakElo.toFixed(2)}</h3>
        </div>
      )}

      <H2>Cumulative Statistics</H2>

      <Table>
        <TableHead>
          <TableRow>
            <TableCell>Category</TableCell>

            <TableCell align="right">Average Time (s)</TableCell>
            <TableCell align="right">Best Time (s)</TableCell>
            <TableCell align="right">Worst Time (s)</TableCell>
            <TableCell align="right">Number completed</TableCell>
          </TableRow>
        </TableHead>

        <TableBody>
          <TableRow>
            <TableCell>Minis (5x5)</TableCell>
            <TableCell align="right">
              {secondsToMinutes(data.MiniStats.Average, true)}
            </TableCell>
            <TableCell align="right">
              {secondsToMinutes(data.MiniStats.Best)}
            </TableCell>
            <TableCell align="right">
              {secondsToMinutes(data.MiniStats.Worst)}
            </TableCell>
            <TableCell align="right">{data.MiniStats.NumCompleted}</TableCell>
          </TableRow>

          <TableRow>
            <TableCell>Saturdays (7x7)</TableCell>
            <TableCell align="right">
              {secondsToMinutes(data.SaturdayStats.Average, true)}
            </TableCell>
            <TableCell align="right">
              {secondsToMinutes(data.SaturdayStats.Best)}
            </TableCell>
            <TableCell align="right">
              {secondsToMinutes(data.SaturdayStats.Worst)}
            </TableCell>
            <TableCell align="right">
              {data.SaturdayStats.NumCompleted}
            </TableCell>
          </TableRow>

          <TableRow>
            <TableCell>Overall</TableCell>
            <TableCell align="right">
              {secondsToMinutes(data.OverallStats.Average, true)}
            </TableCell>
            <TableCell align="right">
              {secondsToMinutes(data.OverallStats.Best)}
            </TableCell>
            <TableCell align="right">
              {secondsToMinutes(data.OverallStats.Worst)}
            </TableCell>
            <TableCell align="right">
              {data.OverallStats.NumCompleted}
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>

      {eloFF && (
        <div css={styles.historyTitle}>
          <H2>Historical Statistics</H2>

          <H2>ELO History</H2>
          <ResponsiveContainer aspect={16 / 9} width={800}>
            <LineChart
              width={730}
              height={250}
              margin={{ top: 20, right: 20, left: 20, bottom: 20 }}
              data={data.EloHistory.slice()
                .reverse()
                .map(({ Date, Elo }) => ({
                  Date: dayjs(Date).utc().unix(),
                  Elo,
                }))}
            >
              <XAxis
                tick={false}
                domain={["auto", "auto"]}
                type="number"
                dataKey="Date"
                label={{ value: "Date", position: "insideBottom" }}
              />
              <YAxis allowDecimals={false} domain={["auto", "auto"]} />

              <Tooltip
                labelFormatter={(value) =>
                  dateToFormat(dayjs.unix(value).utc().toString())
                }
                formatter={(elo: number) => [elo.toFixed(2), "ELO"]}
              />

              <Line dot={false} type="monotone" dataKey="Elo" />
            </LineChart>
          </ResponsiveContainer>

          <Table>
            <TableHead>
              <TableRow>
                <TableCell>Date</TableCell>

                <TableCell align="right">ELO</TableCell>
              </TableRow>
            </TableHead>

            <TableBody>
              {data.EloHistory.map(({ Date, Elo }) => (
                <TableRow key={Date + Elo.toString()}>
                  <TableCell>{dateToFormat(Date)}</TableCell>

                  <TableCell align="right">{Elo.toFixed(2)}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}
    </div>
  );
}
