import { jsx } from "@emotion/react";
import { useState } from "react";
import * as s from "superstruct";
import { datesToFormat, padRight, secondsToMinutes } from "utils";
import Link from "next/link";
import { UserType } from "pages/index";
import * as styles from "./table_styles";
import { useFeatureFlag } from "hooks/use_feature_flag";

interface TableProps {
  daysOfTheWeek: Array<string>;
  rows: Array<s.Infer<typeof UserType>>;
}

interface RowProps {
  UserId: number;
  Username: string;
  WeeksTimes: Array<number>;
  WeeksAverage: number;
  Elo: number;
  shouldUseElo: boolean;
}

function Row({
  UserId,
  Username,
  WeeksTimes,
  WeeksAverage,
  Elo,
  shouldUseElo,
}: RowProps) {
  return (
    <tr css={styles.tableRow}>
      <td>
        <Link href={`/users/${UserId}/`} legacyBehavior>
          <a>{Username}</a>
        </Link>
      </td>

      {padRight(WeeksTimes, -1, 7).map((weeksTime, i) => (
        <td key={Username + weeksTime + i}>
          {weeksTime === -1 ? "" : secondsToMinutes(weeksTime)}
        </td>
      ))}

      <td>{secondsToMinutes(WeeksAverage)}</td>

      {shouldUseElo && <td>{Elo.toFixed(0)}</td>}
    </tr>
  );
}

export const Table = ({ daysOfTheWeek, rows }: TableProps) => {
  const [{ orderBy, ascending }, setOrder] = useState<{
    orderBy: keyof s.Infer<typeof UserType> | number;
    ascending: boolean;
  }>({
    orderBy: "WeeksAverage",
    ascending: true,
  });

  let users = rows.slice();

  let dates = datesToFormat(daysOfTheWeek);

  const [removedUsers, _] = useState(() => {
    if (typeof window !== "undefined") {
      let parsed = window.localStorage.getItem("removedUsers");

      if (!parsed) {
        return [];
      }

      return JSON.parse(parsed);
    }

    return [];
  });

  let newUsers = users
    .filter((user) => !removedUsers.includes(user.Username))
    .filter((user) => user.WeeksTimes.length > 0)
    .map((user) => {
      let newObj: typeof user & { [key: number]: number } = { ...user };
      for (let i = 0; i < user.WeeksTimes.length; ++i) {
        newObj[i] = user.WeeksTimes[i];
      }

      return newObj;
    });

  newUsers.sort((user1, user2) => {
    let user1data =
      user1[orderBy] === -1 || user1[orderBy] === undefined
        ? 10000
        : user1[orderBy];
    let user2data =
      user2[orderBy] === -1 || user2[orderBy] === undefined
        ? 10000
        : user2[orderBy];

    let diff;

    if (typeof user1data === "string" && typeof user2data === "string") {
      diff = user1data.localeCompare(user2data);
    } else if (typeof user1data === "number" && typeof user2data === "number") {
      diff = user1data - user2data;
    } else {
      diff = 0;
    }

    if (ascending) {
      return diff;
    }

    return -diff;
  });

  const { loading, status, error } = useFeatureFlag("elos");

  if (loading) {
    return <div />;
  }

  if (error) {
    return <div>Error</div>;
  }

  const headers: Array<{
    title: string;
    key: keyof s.Infer<typeof UserType> | number;
  }> = [
    { title: "Name", key: "Username" },
    ...dates.map((date, i) => ({ title: date, key: i })),
    { title: "Weekly Average", key: "WeeksAverage" },
  ];

  if (status === true) {
    headers.push({ title: "ELO", key: "Elo" });
  }

  return (
    <table css={styles.table}>
      <thead>
        <tr>
          {headers.map((header, i) => (
            <th
              css={styles.th}
              key={JSON.stringify(header)}
              onClick={() => {
                if (orderBy === header.key) {
                  setOrder({ orderBy, ascending: !ascending });
                } else {
                  setOrder({ orderBy: header.key, ascending: true });
                }
              }}
            >
              {header.title}
              {header.key === orderBy && (
                <div css={styles.arrow}>
                  {!ascending ? " \u25BC" : " \u25B2"}
                </div>
              )}
            </th>
          ))}
        </tr>
      </thead>

      <tbody css={styles.tbody}>
        {newUsers.map((user, i) => (
          <Row {...user} shouldUseElo={status} key={JSON.stringify(user)} />
        ))}
      </tbody>
    </table>
  );
};
