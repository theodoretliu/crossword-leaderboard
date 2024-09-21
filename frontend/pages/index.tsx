import React, { useState } from "react";
import { jsx } from "@emotion/react";
import { Header } from "@/components/header";
import { API_URL } from "api";
import dayjs from "dayjs";
import utc from "dayjs/plugin/utc";
import useSWR from "swr";
import * as z from "zod";
import { Table } from "@/components/table";
import { datesToFormat } from "utils";
import { Button } from "@/components/ui/button";
import { ChevronLeft } from "lucide-react";
import Link from "next/link";
import { TooltipProvider } from "@/components/ui/tooltip";

dayjs.extend(utc);

export const UserType = z.object({
  UserId: z.number(),
  Username: z.string(),
  WeeksTimes: z.array(z.number()),
  WeeksAverage: z.number(),
  Elo: z.number(),
  Qualified: z.boolean(),
});

export const ResponseType = z.object({
  Users: z.array(UserType),
  DaysOfTheWeek: z.array(z.string()),
});

async function fetcher(key: string) {
  const res = await fetch(API_URL + key);
  const json = await res.json();

  return ResponseType.parse(json);
}

export async function getServerSideProps() {
  const initialData = await fetcher("/new");

  return { props: { initialData } };
}

function App({ initialData }: { initialData: z.infer<typeof ResponseType> }) {
  const { error, data } = useSWR("/new", fetcher, {
    fallbackData: initialData,
    refreshInterval: 10 * 1000,
  });

  if (error) {
    return (
      <div className="fixed bottom-0 left-0 right-0 top-0 flex items-center justify-center">
        Something went wrong :(
      </div>
    );
  }

  if (!data) {
    return <Header />;
  }

  let dates = datesToFormat(data.DaysOfTheWeek);

  return (
    <React.Fragment>
      <Header />

      <Table daysOfTheWeek={dates} rows={data.Users} />
    </React.Fragment>
  );
}

export default App;
