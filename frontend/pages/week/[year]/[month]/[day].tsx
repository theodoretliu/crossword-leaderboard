import { jsx } from "@emotion/react";
import { useRouter } from "next/router";
import dayjs from "dayjs";
import utc from "dayjs/plugin/utc";
import { GetServerSideProps } from "next";
import { Header } from "components/header";
import { API_URL } from "api";
import { ResponseType } from "pages/index";
import * as s from "superstruct";
import { Table } from "components/table";
import { H2 } from "components/h2";

import { datesToFormat } from "@/utils";

dayjs.extend(utc);

async function fetcher(key: string) {
  const res = await fetch(API_URL + key);
  const json = await res.json();

  s.assert(json, ResponseType);

  return json;
}

export const getServerSideProps: GetServerSideProps = async (context) => {
  const { year, month, day } = context.query;
  const initialData = await fetcher(`/week/${year}/${month}/${day}`);

  return { props: { initialData } };
};

export default function Week({
  initialData,
}: {
  initialData: s.Infer<typeof ResponseType>;
}) {
  const router = useRouter();

  const { year, month, day } = router.query as {
    year: string;
    month: string;
    day: string;
  };

  const date = dayjs.utc(
    Date.UTC(parseInt(year), parseInt(month) - 1, parseInt(day))
  );

  let dates = datesToFormat(initialData.DaysOfTheWeek);

  return (
    <div>
      <Header />

      <h2 className="text-lg font-semibold px-4 pb-4">
        Week of {date.format("dddd, MMMM D, YYYY")}
      </h2>

      <Table daysOfTheWeek={dates} rows={initialData.Users} />
    </div>
  );
}
