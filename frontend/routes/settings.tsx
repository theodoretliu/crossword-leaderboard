import React, { useEffect } from "react";
import { Header } from "@/components/header";
import useSWR from "swr";
import * as z from "zod";
import { API_URL } from "@/api";
import { Button } from "@/components/ui/button";
import { useRemovedUsers } from "@/hooks/use_removed_users";
import { DEFAULT_TITLE } from "@/router";

const UsersList = z.object({
  Users: z.array(z.object({ Id: z.number(), Name: z.string() })),
});

async function fetcher(key: string) {
  const response = await fetch(API_URL + key);
  const json = await response.json();

  return UsersList.parse(json);
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

export function Settings() {
  const [removedUsers, setRemovedUsers] = useRemovedUsers();

  useEffect(() => {
    document.title = "Mini Crossword Leaderboard: Settings";
    return () => {
      document.title = DEFAULT_TITLE;
    };
  }, []);

  const { error, data } = useSWR("/all_users", fetcher);

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
