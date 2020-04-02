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

function Cell(props) {
  const { row, column, additionalCSS, gray, children } = props;

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

  return <div css={[style, gray && grayStyle, additionalCSS]}>{children}</div>;
}

function Row(props) {
  let {
    user: { name, weeksTimes, weeklyAverage },
    rowNum,
    gray
  } = props;

  return (
    <React.Fragment>
      <Cell row={rowNum + 1} column={1} gray={gray}>
        {name}
      </Cell>
      {weeksTimes.map((weeksTime, i) => {
        return (
          <Cell row={rowNum + 1} column={i + 2} gray={gray}>
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
        <Cell
          row={1}
          column={i + 1}
          additionalCSS={css`
            font-weight: bold;
            color: white;
            background-color: #4d88f8;
          `}
        >
          {header}
        </Cell>
      ))}
      {users.map((user, i) => (
        <Row user={user} rowNum={i + 1} gray={i % 2 === 1} />
      ))}
    </div>
  );
}

export default App;
