import { jsx } from "@emotion/react";
import { useRouter } from "next/router";
import dayjs from "dayjs";
import utc from "dayjs/plugin/utc";
import { GetServerSideProps } from "next";
import { Header } from "components/header";
import { API_URL } from "api";
import * as z from "zod";
import { H2 } from "components/h2";
import useSWR from "swr";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { dateToFormat, secondsToMinutes } from "utils";

const StatsType = z.object({
  Average: z.number(),
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
});

dayjs.extend(utc);

async function fetcher(key: string) {
  const res = await fetch(API_URL + key);
  const json = await res.json();

  return ResponseType.parse(json);
}

export const getServerSideProps: GetServerSideProps = async (context) => {
  const { userId } = context.query;
  const initialData = await fetcher(`/users/${userId}`);

  return { props: { initialData } };
};

export default function User({
  initialData,
}: {
  initialData: z.infer<typeof ResponseType>;
}) {
  const router = useRouter();

  const { userId } = router.query;

  const { error, data } = useSWR(`/users/${userId}`, fetcher, {
    fallbackData: initialData,
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

      <div className="flex flex-col gap-4">
        <h2 className="text-lg font-semibold px-4">
          Statistics for {data.Username}
        </h2>

        <div className="flex flex-row items-center gap-4 px-4">
          <h3>Current Streak: {data.CurrentStreak}</h3>

          <h3>Longest Streak: {data.LongestStreak}</h3>
        </div>

        <h2 className="text-lg font-semibold px-4">Cumulative Statistics</h2>

        <div className="w-full overflow-scroll">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Category</TableHead>

                <TableHead>Average Time (s)</TableHead>
                <TableHead>Best Time (s)</TableHead>
                <TableHead>Worst Time (s)</TableHead>
                <TableHead>Number completed</TableHead>
              </TableRow>
            </TableHeader>

            <TableBody>
              <TableRow>
                <TableCell>Minis (5x5)</TableCell>
                <TableCell>
                  {secondsToMinutes(data.MiniStats.Average, true)}
                </TableCell>
                <TableCell>{secondsToMinutes(data.MiniStats.Best)}</TableCell>
                <TableCell>{secondsToMinutes(data.MiniStats.Worst)}</TableCell>
                <TableCell>{data.MiniStats.NumCompleted}</TableCell>
              </TableRow>

              <TableRow>
                <TableCell>Saturdays (7x7)</TableCell>
                <TableCell>
                  {secondsToMinutes(data.SaturdayStats.Average, true)}
                </TableCell>
                <TableCell>
                  {secondsToMinutes(data.SaturdayStats.Best)}
                </TableCell>
                <TableCell>
                  {secondsToMinutes(data.SaturdayStats.Worst)}
                </TableCell>
                <TableCell>{data.SaturdayStats.NumCompleted}</TableCell>
              </TableRow>

              <TableRow>
                <TableCell>Overall</TableCell>
                <TableCell>
                  {secondsToMinutes(data.OverallStats.Average, true)}
                </TableCell>
                <TableCell>
                  {secondsToMinutes(data.OverallStats.Best)}
                </TableCell>
                <TableCell>
                  {secondsToMinutes(data.OverallStats.Worst)}
                </TableCell>
                <TableCell>{data.OverallStats.NumCompleted}</TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}
