/** @jsx jsx */
import { jsx } from "@emotion/core";
import { useState } from "react";
import * as t from "io-ts";
import { padRight, secondsToMinutes } from "utils";
import { UserType } from "pages/index";
import * as styles from "./table_styles";

interface TableProps {
  daysOfTheWeek: Array<string>;
  rows: Array<t.TypeOf<typeof UserType>>;
}

interface RowProps {
  Username: string;
  WeeksTimes: Array<number>;
  WeeksAverage: number;
}

function Row({ Username, WeeksTimes, WeeksAverage }: RowProps) {
  return (
    <tr css={styles.tableRow}>
      <td>{Username}</td>

      {padRight(WeeksTimes, -1, 7).map((weeksTime, i) => (
        <td key={Username + weeksTime + i}>
          {weeksTime === -1 ? "" : secondsToMinutes(weeksTime)}
        </td>
      ))}

      <td>{secondsToMinutes(WeeksAverage)}</td>
    </tr>
  );
}

export const Table = ({ daysOfTheWeek, rows }: TableProps) => {
  const [{ orderBy, ascending }, setOrder] = useState<{
    orderBy: keyof t.TypeOf<typeof UserType> | number;
    ascending: boolean;
  }>({
    orderBy: "WeeksAverage",
    ascending: true,
  });

  let users = rows.slice();

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

  const headers: Array<{
    title: string;
    key: keyof t.TypeOf<typeof UserType> | number;
  }> = [
    { title: "Name", key: "Username" },
    ...daysOfTheWeek.map((date, i) => ({ title: date, key: i })),
    { title: "Weekly Average", key: "WeeksAverage" },
  ];

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
          <Row {...user} key={JSON.stringify(user)} />
        ))}
      </tbody>
    </table>
  );
};
