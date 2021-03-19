/** @jsx jsx */
import { jsx } from "@emotion/core";
import { useRouter } from "next/router";
import dayjs from "dayjs";
import utc from "dayjs/plugin/utc";
import { GetServerSideProps } from "next";
import { Header } from "components/header";
import { API_URL } from "api";
import { ResponseType } from "pages/index";
import * as t from "io-ts";
import { Table } from "components/table";

import * as styles from "./[day]_styles";

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
  const { year, month, day } = context.query;
  const initialData = await fetcher(`/week/${year}/${month}/${day}`);

  return { props: { initialData } };
};

export default function Week({
  initialData,
}: {
  initialData: t.TypeOf<typeof ResponseType>;
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

  return (
    <div>
      <Header />
      <h2 css={styles.weekTitle}>
        Week of {date.format("dddd, MMMM D, YYYY")}
      </h2>
      <Table
        daysOfTheWeek={initialData.DaysOfTheWeek}
        rows={initialData.Users}
      />
    </div>
  );
}
