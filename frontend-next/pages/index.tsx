/** @jsx jsx */
import React, { useState } from "react";
import { css, jsx } from "@emotion/core";
import { Header } from "components/header";
import { secondsToMinutes, withApollo } from "../utils";
import { API_URL } from "api";
import dayjs from "dayjs";
import utc from "dayjs/plugin/utc";
import useSWR from "swr";
import * as t from "io-ts";

dayjs.extend(utc);

const UserType = t.type({
  Username: t.string,
  WeeksTimes: t.array(t.number),
  WeeksAverage: t.number,
});

const ResponseType = t.type({
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

export const tableStyle = css`
  display: grid;
  grid-template-columns: repeat(8, auto);
  width: 100%;
  height: auto;
`;

interface CellProps {
  onClick?: () => void;
  row: number;
  column: number;
  additionalCSS?: ReturnType<typeof css>;
  gray?: boolean;
  children: React.ReactNode;
}

function Cell(props: CellProps) {
  const { onClick, row, column, additionalCSS, gray, children } = props;

  const style = css`
    grid-row-start: ${row};
    grid-column-start: ${column};
    padding: 20px;
    font-family: "Roboto", sans-serif;
    display: flex;
    flex-direction: row;
    align-items: center;
    justify-content: center;
    text-align: center;
  `;

  const grayStyle = css`
    background-color: #e0e0e0;
  `;

  return (
    <div onClick={onClick} css={[style, gray && grayStyle, additionalCSS]}>
      {children}
    </div>
  );
}

function padRight<T>(arr: Array<T>, value: T, length: number): Array<T> {
  let newArr = arr.slice();

  while (newArr.length < length) {
    newArr.push(value);
  }

  return newArr;
}

interface RowProps {
  User: { Username: string; WeeksTimes: Array<number>; WeeksAverage: number };
  rowNum: number;
  gray: boolean;
}
function Row({
  User: { Username, WeeksTimes, WeeksAverage },
  rowNum,
  gray,
}: RowProps) {
  return (
    <React.Fragment>
      <Cell row={rowNum + 1} column={1} gray={gray}>
        {Username}
      </Cell>
      {padRight(WeeksTimes, -1, 7).map((weeksTime, i) => {
        return (
          <Cell
            key={Username + weeksTime + i}
            row={rowNum + 1}
            column={i + 2}
            gray={gray}
          >
            {weeksTime === -1 ? "" : secondsToMinutes(weeksTime)}
          </Cell>
        );
      })}
      <Cell row={rowNum + 1} column={9} gray={gray}>
        {secondsToMinutes(WeeksAverage)}
      </Cell>
    </React.Fragment>
  );
}

function App({ initialData }: { initialData: t.TypeOf<typeof ResponseType> }) {
  const [{ orderBy, ascending }, setOrder] = useState<{
    orderBy: keyof t.TypeOf<typeof UserType> | number;
    ascending: boolean;
  }>({
    orderBy: "WeeksAverage",
    ascending: true,
  });

  // eslint-disable-next-line
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

  let dates = data.DaysOfTheWeek.map((x) =>
    dayjs(x).utc().format("dddd, MMMM D, YYYY")
  );

  let users = data.Users.slice();

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
    ...dates.map((date, i) => ({ title: date, key: i })),
    { title: "Weekly Average", key: "WeeksAverage" },
  ];

  return (
    <React.Fragment>
      <Header />
      <div css={tableStyle}>
        {headers.map((header, i) => (
          <Cell
            key={JSON.stringify(header)}
            row={1}
            column={i + 1}
            additionalCSS={css`
              font-weight: bold;
              color: white;
              background-color: #4d88f8;
              cursor: pointer;
              position: relative;
            `}
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
              <div
                css={css`
                  position: absolute;
                  font-size: 12px;
                  right: 5px;
                  bottom: 5px;
                `}
              >
                {!ascending ? " \u25BC" : " \u25B2"}
              </div>
            )}
          </Cell>
        ))}
        {newUsers.map((user, i) => (
          <Row
            key={JSON.stringify(user)}
            User={user}
            rowNum={i + 1}
            gray={i % 2 === 1}
          />
        ))}
      </div>
    </React.Fragment>
  );
}

export default App;
