/** @jsx jsx */
import { jsx, css } from "@emotion/core";
import React, { useState, useEffect } from "react";
import { Header } from "./Header";
import { gql } from "apollo-boost";
import { useQuery } from "@apollo/react-hooks";

const GET_USERS = gql`
  query GetUsers {
    users {
      name
    }
  }
`;

function UserRow(props) {
  const { name, onClick } = props;

  return (
    <div
      css={css`
        display: block;
        padding: 5px 0px;
      `}
    >
      <span
        css={css`
          cursor: pointer;
        `}
        onClick={onClick}
      >
        {name}
      </span>
    </div>
  );
}
export function Settings() {
  const [removedUsers, setRemovedUsers] = useState(() => {
    let parsed = window.localStorage.getItem("removedUsers");

    if (!parsed) {
      return [];
    }

    return JSON.parse(parsed);
  });

  useEffect(() => {
    window.localStorage.setItem("removedUsers", JSON.stringify(removedUsers));
  }, [removedUsers]);

  const { loading, error, data } = useQuery(GET_USERS);

  if (loading) {
    return <Header />;
  }

  if (error) {
    return (
      <React.Fragment>
        <Header />
        {"There was an error."}
      </React.Fragment>
    );
  }

  let users = data.users;

  let allowedUsers = users.filter((user) => !removedUsers.includes(user.name));
  allowedUsers.sort((u1, u2) => u1.name.localeCompare(u2.name));

  let blockedUsers = users.filter((user) => removedUsers.includes(user.name));

  blockedUsers.sort((u1, u2) => u1.name.localeCompare(u2.name));

  const buttonStyle = css`
    padding: 20px;
    font-weight: bold;
    color: white;
    background-color: #4d88f8;
    border-radius: 5px;
    width: fit-content;
    cursor: pointer;
    margin-bottom: 10px;
  `;

  return (
    <React.Fragment>
      <Header />
      <div
        css={css`
          display: grid;
          width: 100%;
          grid-template-columns: 1fr 1fr;
          font-family: Roboto, sans-serif;
        `}
      >
        <div
          css={css`
            padding: 0px 20px 20px 20px;
            grid-column-start: 1;
          `}
        >
          <h2>Shown (Click to remove)</h2>
          <div
            css={buttonStyle}
            onClick={() => setRemovedUsers(users.map((user) => user.name))}
          >
            Remove all
          </div>
          {allowedUsers.map((user) => (
            <UserRow
              key={user.name}
              name={user.name}
              onClick={() => setRemovedUsers([...removedUsers, user.name])}
            />
          ))}
        </div>

        <div
          css={css`
            padding: 0px 20px 20px 20px;
            grid-column-start: 2;
          `}
        >
          <h2>Hidden (Click to restore)</h2>
          <div css={buttonStyle} onClick={() => setRemovedUsers([])}>
            Restore all
          </div>
          {blockedUsers.map((user) => (
            <UserRow
              key={user.name}
              name={user.name}
              onClick={() =>
                setRemovedUsers(
                  removedUsers.filter((username) => username !== user.name)
                )
              }
            />
          ))}
        </div>
      </div>
    </React.Fragment>
  );
}
