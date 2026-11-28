import React from "react";
import { Header } from "@/components/header";
import { API_URL } from "@/api";
import dayjs from "dayjs";
import utc from "dayjs/plugin/utc";
import useSWR from "swr";
import { ResponseType } from "@/lib/schemas";
import { Table } from "@/components/table";
import { datesToFormat } from "@/utils";

dayjs.extend(utc);

async function fetcher(key: string) {
  const res = await fetch(API_URL + key);
  const json = await res.json();

  return ResponseType.parse(json);
}

export function Index() {
  const { error, data } = useSWR("/new", fetcher, {
    refreshInterval: 10 * 1000,
  });

  if (error) {
    return (
      <div className="fixed top-0 right-0 bottom-0 left-0 flex items-center justify-center">
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
