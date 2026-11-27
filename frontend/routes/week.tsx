import { getRouteApi } from "@tanstack/react-router";
import dayjs from "dayjs";
import utc from "dayjs/plugin/utc";
import useSWR from "swr";
import { Header } from "@/components/header";
import { API_URL } from "@/api";
import { ResponseType } from "@/lib/schemas";
import { Table } from "@/components/table";
import { datesToFormat } from "@/utils";

dayjs.extend(utc);

async function fetcher(key: string) {
  const res = await fetch(API_URL + key);
  const json = await res.json();

  return ResponseType.parse(json);
}

const route = getRouteApi("/week/$year/$month/$day");

export function Week() {
  const { year, month, day } = route.useParams();

  const { error, data } = useSWR(`/week/${year}/${month}/${day}`, fetcher);

  if (error) {
    return "Error";
  }

  if (!data) {
    return <Header />;
  }

  const date = dayjs.utc(Date.UTC(parseInt(year), parseInt(month) - 1, parseInt(day)));

  let dates = datesToFormat(data.DaysOfTheWeek);

  return (
    <div>
      <Header />

      <h2 className="px-4 pb-4 text-lg font-semibold">
        Week of {date.format("dddd, MMMM D, YYYY")}
      </h2>

      <Table daysOfTheWeek={dates} rows={data.Users} />
    </div>
  );
}
