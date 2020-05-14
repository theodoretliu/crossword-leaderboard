/** @jsx jsx */
import React, { useState } from "react";
import { css, jsx } from "@emotion/core";
import { getDataFromTree } from "@apollo/react-ssr";
import { ApolloProvider, useQuery } from "@apollo/react-hooks";
import { gql } from "apollo-boost";
import { Header } from "../components/Header";
import dayjs from "dayjs";
import utc from "dayjs/plugin/utc";

import { ApolloClient } from "apollo-client";
import { createHttpLink } from "apollo-link-http";
import { StaticRouter as Router } from "react-router-dom";
import { InMemoryCache } from "apollo-cache-inmemory";

dayjs.extend(utc);

const GET_DATA = gql`
  query GetData {
    daysOfTheWeek

    users {
      weeksTimes
      weeklyAverage
      name
    }
  }
`;

export const tableStyle = css`
  display: grid;
  grid-template-columns: repeat(8, auto);
  width: 100%;
  height: auto;
`;

function Cell(props) {
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

function Row(props) {
  let {
    user: { name, weeksTimes, weeklyAverage },
    rowNum,
    gray,
  } = props;

  return (
    <React.Fragment>
      <Cell row={rowNum + 1} column={1} gray={gray}>
        {name}
      </Cell>
      {weeksTimes.map((weeksTime, i) => {
        return (
          <Cell
            key={name + weeksTime + i}
            row={rowNum + 1}
            column={i + 2}
            gray={gray}
          >
            {weeksTime === -1 ? "-" : weeksTime}
          </Cell>
        );
      })}
      <Cell row={rowNum + 1} column={9} gray={gray}>
        {weeklyAverage}
      </Cell>
    </React.Fragment>
  );
}

function App() {
  const [{ orderBy, ascending }, setOrder] = useState({
    orderBy: "weeklyAverage",
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

  const { loading, error, data } = useQuery(GET_DATA, {
    pollInterval: 10 * 1000,
  });

  if (loading) {
    return (
      <React.Fragment>
        <Header />
        {"Loading"}
      </React.Fragment>
    );
  }

  if (error) {
    return "there was an error";
  }

  let dates = data.daysOfTheWeek.map((x) =>
    dayjs(x).utc().format("dddd, MMMM D, YYYY")
  );

  let users = data.users.slice();

  users = users
    .filter((user) => !removedUsers.includes(user.name))
    .map((user) => {
      let newObj = { ...user };
      for (let i = 0; i < 7; ++i) {
        newObj[i] = user.weeksTimes[i];
      }

      return newObj;
    });

  users.sort((user1, user2) => {
    let user1data = user1[orderBy] === -1 ? 10000 : user1[orderBy];
    let user2data = user2[orderBy] === -1 ? 10000 : user2[orderBy];

    let diff;

    if (typeof user1data === "string") {
      diff = user1data.localeCompare(user2data);
    } else {
      diff = user1data - user2data;
    }

    if (ascending) {
      return diff;
    }

    return -diff;
  });

  const headers = [
    { title: "Name", key: "name" },
    ...dates.map((date, i) => ({ title: date, key: i })),
    { title: "Weekly Average", key: "weeklyAverage" },
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
        {users.map((user, i) => (
          <Row
            key={JSON.stringify(user)}
            user={user}
            rowNum={i + 1}
            gray={i % 2 === 1}
          />
        ))}
      </div>
    </React.Fragment>
  );
}

export async function getServerSideProps(context) {
  let client = new ApolloClient({
    ssrMode: true,
    link: createHttpLink({
      uri: process.env.NEXT_PUBLIC_API_URL || "http://localhost:8000/graphql",
      credentials: "same-origin",
    }),
    cache: new InMemoryCache(),
  });

  const c = {};

  const A = (
    <ApolloProvider client={client}>
      <App />
    </ApolloProvider>
  );

  await getDataFromTree(A);

  return { props: client.extract() };
}

export default function (props) {
  return <App />;
}