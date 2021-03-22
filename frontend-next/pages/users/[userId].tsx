/** @jsx jsx */
import { jsx } from "@emotion/core";
import { useRouter } from "next/router";
import dayjs from "dayjs";
import utc from "dayjs/plugin/utc";
import { GetServerSideProps } from "next";
import { Header } from "components/header";
import { API_URL } from "api";
import * as t from "io-ts";
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
import { secondsToMinutes } from "utils";

const StatsType = t.type({
  Average: t.number,
  Best: t.number,
  Worst: t.number,
  NumCompleted: t.number,
});

const ResponseType = t.type({
  Username: t.string,
  MiniStats: StatsType,
  SaturdayStats: StatsType,
  OverallStats: StatsType,
});

dayjs.extend(utc);

async function fetcher(key: string) {
  const res = await fetch(API_URL + key);
  const json = await res.json();

  const decoded = ResponseType.decode(json);

  if (decoded._tag == "Left") {
    throw decoded.left;
  }

  return decoded.right;
}

export const getServerSideProps: GetServerSideProps = async (context) => {
  const { userId } = context.query;
  const initialData = await fetcher(`/users/${userId}`);

  return { props: { initialData } };
};

export default function User({
  initialData,
}: {
  initialData: t.TypeOf<typeof ResponseType>;
}) {
  const router = useRouter();

  const { userId } = router.query;

  const { error, data } = useSWR(`/users/${userId}`, fetcher, { initialData });

  if (error) {
    return <div>hello</div>;
  }

  if (!data) {
    return <div />;
  }

  return (
    <div>
      <Header />
      <H2>{data.Username}</H2>

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
    </div>
  );
}
