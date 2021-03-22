/** @jsx jsx */
import React, { useState } from "react";
import { jsx } from "@emotion/core";
import { Header } from "components/header";
import { API_URL } from "api";
import dayjs from "dayjs";
import utc from "dayjs/plugin/utc";
import useSWR from "swr";
import * as t from "io-ts";
import { Table } from "components/table";
import { datesToFormat } from "utils";

dayjs.extend(utc);

export const UserType = t.type({
  UserId: t.number,
  Username: t.string,
  WeeksTimes: t.array(t.number),
  WeeksAverage: t.number,
  Elo: t.number,
});

export const ResponseType = t.type({
  Users: t.array(UserType),
  DaysOfTheWeek: t.array(t.string),
});

async function fetcher(key: string) {
  const res = await fetch(API_URL + key);
  const json = await res.json();

  const decoded = ResponseType.decode(json);

  if (decoded._tag == "Left") {
    throw decoded.left;
  }

  return decoded.right;
}

export async function getServerSideProps() {
  const initialData = await fetcher("/new");

  return { props: { initialData } };
}

function App({ initialData }: { initialData: t.TypeOf<typeof ResponseType> }) {
  const { error, data } = useSWR("/new", fetcher, {
    initialData,
    refreshInterval: 10 * 1000,
  });

  if (error) {
    return <div>hello</div>;
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
