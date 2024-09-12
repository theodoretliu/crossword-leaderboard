import React from "react";
import useSWR from "swr";
import { API_URL } from "api";
import * as z from "zod";
import { Header } from "@/components/header";

const LeaderboardEntrySchema = z.object({
  ID: z.number(),
  Name: z.string(),
  Percentile10: z.number(),
  Percentile20: z.number(),
  Percentile30: z.number(),
  Percentile40: z.number(),
  Percentile50: z.number(),
  Percentile60: z.number(),
  Percentile70: z.number(),
  Percentile80: z.number(),
  Percentile90: z.number(),
});

const LeaderboardSchema = z.array(LeaderboardEntrySchema);

type LeaderboardEntry = z.infer<typeof LeaderboardEntrySchema>;

async function fetcher(url: string) {
  const res = await fetch(url);
  const data = await res.json();
  return LeaderboardSchema.parse(data);
}

export async function getServerSideProps() {
  const initialData = await fetcher(API_URL + "/leaderboard");
  return { props: { initialData } };
}

const Leaderboard: React.FC<{ initialData: LeaderboardEntry[] }> = ({
  initialData,
}) => {
  const { data: leaderboardData } = useSWR<LeaderboardEntry[]>(
    API_URL + "/leaderboard",
    fetcher,
    { fallbackData: initialData, refreshInterval: 10000 }
  );

  if (!leaderboardData) {
    return <div>Loading...</div>;
  }

  return (
    <div>
      <Header />

      <h1>Leaderboard</h1>

      <table>
        <thead>
          <tr>
            <th rowSpan={1}></th>
            <th colSpan={9}>Percentiles</th>
          </tr>
          <tr>
            <th>Name</th>
            <th>10</th>
            <th>20</th>
            <th>30</th>
            <th>40</th>
            <th>50</th>
            <th>60</th>
            <th>70</th>
            <th>80</th>
            <th>90</th>
          </tr>
        </thead>
        <tbody>
          {leaderboardData?.map((entry) => (
            <tr key={entry.ID}>
              <td>{entry.Name}</td>
              <td title={entry.Percentile10.toFixed(2)}>
                {Math.round(entry.Percentile10)}
              </td>
              <td title={entry.Percentile20.toFixed(2)}>
                {Math.round(entry.Percentile20)}
              </td>
              <td title={entry.Percentile30.toFixed(2)}>
                {Math.round(entry.Percentile30)}
              </td>
              <td title={entry.Percentile40.toFixed(2)}>
                {Math.round(entry.Percentile40)}
              </td>
              <td title={entry.Percentile50.toFixed(2)}>
                {Math.round(entry.Percentile50)}
              </td>
              <td title={entry.Percentile60.toFixed(2)}>
                {Math.round(entry.Percentile60)}
              </td>
              <td title={entry.Percentile70.toFixed(2)}>
                {Math.round(entry.Percentile70)}
              </td>
              <td title={entry.Percentile80.toFixed(2)}>
                {Math.round(entry.Percentile80)}
              </td>
              <td title={entry.Percentile90.toFixed(2)}>
                {Math.round(entry.Percentile90)}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};

export default Leaderboard;
