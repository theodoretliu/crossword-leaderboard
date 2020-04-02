/** @jsx jsx */
import React from "react";
import { css, jsx } from "@emotion/core";
import { useQuery } from "@apollo/react-hooks";
import { gql } from "apollo-boost";
import dayjs from "dayjs";
import utc from "dayjs/plugin/utc";

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

function Row(props) {
  let {
    user: { name, weeksTimes, weeklyAverage },
    rowNum
  } = props;

  return (
    <React.Fragment>
      <div
        css={css`
          grid-row-start: ${rowNum + 1};
          grid-column-start: 1;
          padding: 20px;
        `}
      >
        {name}
      </div>
      {weeksTimes.map((weeksTime, i) => {
        return (
          <div
            css={css`
              grid-row-start: ${rowNum + 1};
              grid-column-start: ${i + 2};
              padding: 20px;
            `}
          >
            {weeksTime === -1 ? "-" : weeksTime}
          </div>
        );
      })}
      <div
        css={css`
          grid-row-start: ${rowNum + 1};
          grid-column-start: 9;
          padding: 20px;
        `}
      >
        {weeklyAverage}
      </div>
    </React.Fragment>
  );
}

function App() {
  const { loading, error, data } = useQuery(GET_DATA);

  if (loading) {
    return "Loading...";
  }

  if (error) {
    return "there was an error";
  }

  let dates = data.daysOfTheWeek.map(x =>
    dayjs(x)
      .utc()
      .format("dddd, MMMM D, YYYY")
  );

  let users = data.users.slice();

  users.sort((user1, user2) => user1.weeklyAverage - user2.weeklyAverage);

  return (
    <div css={tableStyle}>
      {["Name", ...dates, "Weekly Average"].map((header, i) => (
        <div
          css={css`
            grid-column-start: ${i + 1};
            grid-row-start: 1;
            border: 1px solid;
            border-right: none;
            border-bottom: none;
            text-align: center;
            padding: 20px;
            font-weight: bold;
            display: flex;
            flex-direction: row;
            align-items: center;
            justify-content: center;
          `}
        >
          {header}
        </div>
      ))}
      {users.map((user, i) => (
        <Row user={user} rowNum={i + 1} />
      ))}
    </div>
  );
}

export default App;
