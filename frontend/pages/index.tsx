import React, { useState } from "react";
import { jsx } from "@emotion/react";
import { Header } from "@/components/header";
import { API_URL } from "api";
import dayjs from "dayjs";
import utc from "dayjs/plugin/utc";
import useSWR from "swr";
import * as s from "superstruct";
import { Table } from "@/components/table";
import { datesToFormat } from "utils";
import { assert } from "superstruct";
import { Button } from "@/components/ui/button";
import { ChevronLeft } from "lucide-react";
import Link from "next/link";

dayjs.extend(utc);

export const UserType = s.object({
  UserId: s.number(),
  Username: s.string(),
  WeeksTimes: s.array(s.number()),
  WeeksAverage: s.number(),
  Elo: s.number(),
});

export const ResponseType = s.object({
  Users: s.array(UserType),
  DaysOfTheWeek: s.array(s.string()),
});

async function fetcher(key: string) {
  const res = await fetch(API_URL + key);
  const json = await res.json();

  assert(json, ResponseType);

  return json;
}

export async function getServerSideProps() {
  const initialData = await fetcher("/new");

  return { props: { initialData } };
}

function App({ initialData }: { initialData: s.Infer<typeof ResponseType> }) {
  const { error, data } = useSWR("/new", fetcher, {
    fallbackData: initialData,
    refreshInterval: 10 * 1000,
  });

  if (error) {
    return (
      <div className="fixed top-0 bottom-0 left-0 right-0 flex items-center justify-center">
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
