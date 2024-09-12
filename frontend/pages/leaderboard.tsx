import React, { useState } from "react";
import useSWR from "swr";
import { API_URL } from "api";
import * as z from "zod";
import { Header } from "@/components/header";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import Link from "next/link";
import { cn } from "@/lib/utils";

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

  const [sortColumn, setSortColumn] =
    useState<keyof LeaderboardEntry>("Percentile50");

  const [ascending, setAscending] = useState<boolean>(true);

  if (!leaderboardData) {
    return <div>Loading...</div>;
  }

  const sortedData = [...leaderboardData].sort((a, b) => {
    if (a[sortColumn] < b[sortColumn]) return ascending ? -1 : 1;
    if (a[sortColumn] > b[sortColumn]) return ascending ? 1 : -1;
    return 0;
  });

  const handleSort = (column: keyof LeaderboardEntry) => {
    if (column === sortColumn) {
      setAscending(!ascending);
    } else {
      setSortColumn(column);
      setAscending(true);
    }
  };

  return (
    <div>
      <Header />

      <h1 className="text-lg font-semibold p-4 pt-0">All-Time Leaderboard</h1>

      <Table>
        <TableHeader className="bg-blue-500 [&_tr]:border-none">
          <TableRow className="hover:bg-nonsense">
            <TableHead colSpan={1} className="h-[30px]"></TableHead>
            <TableHead colSpan={9} className="text-center text-white h-[30px]">
              Percentiles
            </TableHead>
          </TableRow>

          <TableRow className="hover:bg-nonsense">
            <TableHead className="text-white">Name</TableHead>
            {[10, 20, 30, 40, 50, 60, 70, 80, 90].map((percentile) => (
              <TableHead
                key={percentile}
                className={cn(
                  "relative p-4 cursor-pointer text-white text-right"
                )}
                onClick={() =>
                  handleSort(
                    `Percentile${percentile}` as keyof LeaderboardEntry
                  )
                }
              >
                <span className="line-clamp-2">{percentile}</span>
                {sortColumn === `Percentile${percentile}` && (
                  <div className="absolute bottom-1 right-1">
                    {ascending ? " \u25B2" : " \u25BC"}
                  </div>
                )}
              </TableHead>
            ))}
          </TableRow>
        </TableHeader>

        <TableBody>
          {sortedData.map((entry) => (
            <TableRow
              key={entry.ID}
              className={
                entry.Name.toLowerCase() === "everyone"
                  ? "bg-blue-100 font-semibold hover:bg-blue-100"
                  : ""
              }
            >
              <TableCell>
                <Link
                  href={`/users/${entry.ID}`}
                  className="text-primary underline-offset-4 hover:underline"
                >
                  {entry.Name}
                </Link>
              </TableCell>
              {[10, 20, 30, 40, 50, 60, 70, 80, 90].map((percentile) => (
                <TableCell
                  key={percentile}
                  title={entry[
                    `Percentile${percentile}` as keyof Omit<
                      LeaderboardEntry,
                      "ID" | "Name"
                    >
                  ].toFixed(2)}
                  className="text-right"
                >
                  {Math.round(
                    entry[
                      `Percentile${percentile}` as keyof Omit<
                        LeaderboardEntry,
                        "ID" | "Name"
                      >
                    ]
                  )}
                </TableCell>
              ))}
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
};

export default Leaderboard;
