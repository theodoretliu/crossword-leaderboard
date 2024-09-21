import { jsx, css } from "@emotion/react";
import React, { useState, useEffect } from "react";
import { Header } from "@/components/header";
import Head from "next/head";
import useSWR from "swr";
import * as z from "zod";
import { API_URL } from "api";
import { Button } from "@/components/ui/button";
import { useRemovedUsers } from "@/hooks/use_removed_users";

const UsersList = z.object({
  Users: z.array(z.object({ Id: z.number(), Name: z.string() })),
});

async function fetcher(key: string) {
  const response = await fetch(API_URL + key);
  const json = await response.json();

  return UsersList.parse(json);
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
    <Button variant="link" onClick={onClick} className="w-fit px-0" size="sm">
      {name}
    </Button>
  );
}

interface SettingsProps {
  initialData: z.infer<typeof UsersList>;
}

function Settings({ initialData }: SettingsProps) {
  const [removedUsers, setRemovedUsers] = useRemovedUsers();

  const [onClient, setIsOnClient] = useState(false);

  useEffect(() => {
    setIsOnClient(true);
  }, []);

  const { error, data } = useSWR("/all_users", fetcher, {
    fallbackData: initialData,
  });

  if (!onClient) {
    return null;
  }

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

  let users = data.Users;

  let allowedUsers = users.filter((user) => !removedUsers.includes(user.Id));
  allowedUsers.sort((u1, u2) => u1.Name.localeCompare(u2.Name));

  let blockedUsers = users.filter((user) => removedUsers.includes(user.Id));

  blockedUsers.sort((u1, u2) => u1.Name.localeCompare(u2.Name));

  return (
    <React.Fragment>
      <Head>
        <title>Mini Crossword Leaderboard: Settings</title>
      </Head>

      <Header />

      <h2 className="px-4 pb-4 text-lg font-semibold">Settings</h2>

      <div className="grid w-full grid-cols-2 gap-4 px-4 pb-4">
        <div className="flex flex-col gap-4">
          <h2>Shown (Click name to remove)</h2>

          <Button onClick={() => setRemovedUsers(users.map((user) => user.Id))} className="w-fit">
            Remove all
          </Button>

          <div className="flex flex-col gap-1">
            {allowedUsers.map((user) => (
              <UserRow
                key={user.Id}
                name={user.Name}
                onClick={() => setRemovedUsers([...removedUsers, user.Id])}
              />
            ))}
          </div>
        </div>

        <div className="flex flex-col gap-4">
          <h2>Hidden (Click name to restore)</h2>

          <Button onClick={() => setRemovedUsers([])} className="w-fit">
            Restore all
          </Button>

          <div className="flex flex-col gap-1">
            {blockedUsers.map((user) => (
              <UserRow
                key={user.Id}
                name={user.Name}
                onClick={() => setRemovedUsers(removedUsers.filter((userId) => userId !== user.Id))}
              />
            ))}
          </div>
        </div>
      </div>
    </React.Fragment>
  );
}

export default Settings;
