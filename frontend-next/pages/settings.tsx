/** @jsx jsx */
import { jsx, css } from "@emotion/react";
import React, { useState, useEffect } from "react";
import { Header } from "components/header";
import Head from "next/head";
import useSWR from "swr";
import { H2 } from "components/h2";
import * as s from "superstruct";
import { API_URL } from "api";

const UsersList = s.object({
  Users: s.array(s.string()),
});

async function fetcher(key: string) {
  const response = await fetch(API_URL + key);
  const json = await response.json();

  s.assert(json, UsersList);

  return json;
}

export async function getServerSideProps() {
  const initialData = await fetcher("/all_users");

  return { props: { initialData } };
}

interface UserRowProps {
  name: string;
  onClick: () => void;
}

function UserRow(props: UserRowProps) {
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

interface SettingsProps {
  initialData: { Users: Array<string> };
}

function Settings({ initialData }: SettingsProps) {
  const [removedUsers, setRemovedUsers] = useState<string[]>(() => {
    if (typeof window === "undefined") {
      return [];
    }

    let parsed = window.localStorage.getItem("removedUsers");

    if (!parsed) {
      return [];
    }

    return JSON.parse(parsed);
  });

  useEffect(() => {
    window.localStorage.setItem("removedUsers", JSON.stringify(removedUsers));
  }, [removedUsers]);

  const { error, data } = useSWR("/all_users", fetcher, {
    initialData,
  });

  if (error) {
    return (
      <React.Fragment>
        <Header />
        {"There was an error."}
      </React.Fragment>
    );
  }

  if (!data) {
    return "loading";
  }

  let usernames = data.Users;

  let allowedUsers = usernames.filter((user) => !removedUsers.includes(user));
  allowedUsers.sort((u1, u2) => u1.localeCompare(u2));

  let blockedUsers = usernames.filter((user) => removedUsers.includes(user));

  blockedUsers.sort((u1, u2) => u1.localeCompare(u2));

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
      <Head>
        <title>Mini Crossword Leaderboard: Settings</title>
      </Head>
      <Header />
      <H2>Settings</H2>
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
            onClick={() => setRemovedUsers(usernames.map((user) => user))}
          >
            Remove all
          </div>
          {allowedUsers.map((user) => (
            <UserRow
              key={user}
              name={user}
              onClick={() => setRemovedUsers([...removedUsers, user])}
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
              key={user}
              name={user}
              onClick={() =>
                setRemovedUsers(
                  removedUsers.filter((username) => username !== user)
                )
              }
            />
          ))}
        </div>
      </div>
    </React.Fragment>
  );
}

export default Settings;
