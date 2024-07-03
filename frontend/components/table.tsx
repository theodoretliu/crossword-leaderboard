import { jsx } from "@emotion/react";
import { useState } from "react";
import * as z from "zod";
import { datesToFormat, padRight, secondsToMinutes } from "utils";
import Link from "next/link";
import { UserType } from "pages/index";
import {
  Table as UITable,
  TableBody,
  TableHead,
  TableHeader,
  TableRow,
  TableCell,
} from "./ui/table";
import { Button } from "./ui/button";
import { cn } from "@/lib/utils";
import _sortBy from "lodash/sortBy";
import _reversed from "lodash/reverse";

interface TableProps {
  daysOfTheWeek: Array<string>;
  rows: Array<z.infer<typeof UserType>>;
}

interface RowProps {
  UserId: number;
  Username: string;
  WeeksTimes: Array<number>;
  WeeksAverage: number;
  Elo: number;
}

function Row({ UserId, Username, WeeksTimes, WeeksAverage, Elo }: RowProps) {
  return (
    <TableRow className="text-right">
      <TableCell className="text-left p-4">
        <Link
          href={`/users/${UserId}/`}
          className="text-primary underline-offset-4 hover:underline"
        >
          {Username}
        </Link>
      </TableCell>

      {padRight(WeeksTimes, -1, 7).map((weeksTime, i) => (
        <TableCell key={Username + weeksTime + i} className="p-4">
          {weeksTime === -1 ? "" : secondsToMinutes(weeksTime)}
        </TableCell>
      ))}

      <TableCell className="p-4">{secondsToMinutes(WeeksAverage)}</TableCell>
    </TableRow>
  );
}

export const Table = ({ daysOfTheWeek, rows }: TableProps) => {
  const [{ orderBy, ascending }, setOrder] = useState<{
    orderBy: keyof z.infer<typeof UserType> | number;
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
    .filter((user) => user.WeeksTimes.length > 0)
    .map((user) => {
      let newObj: typeof user & { [key: number]: number } = { ...user };
      for (let i = 0; i < user.WeeksTimes.length; ++i) {
        newObj[i] = user.WeeksTimes[i];
      }

      return newObj;
    });

  let sortedUsers = (() => {
    let prelimSort;
    if (orderBy === "WeeksAverage") {
      prelimSort = _sortBy(newUsers, [
        (user) => (user.Qualified ? 0 : 1),
        "WeeksAverage",
      ]);
    } else if (orderBy === "Username") {
      prelimSort = _sortBy(newUsers, orderBy);
    } else {
      prelimSort = _sortBy(newUsers, (user) =>
        user[orderBy] === -1 ? 10000 : user[orderBy]
      );
    }

    if (ascending) {
      return prelimSort;
    }

    return _reversed(prelimSort);
  })();

  const headers: Array<{
    title: string;
    key: keyof z.infer<typeof UserType> | number;
  }> = [
    { title: "Name", key: "Username" },
    ...daysOfTheWeek.map((date, i) => ({ title: date, key: i })),
    { title: "Weekly Average", key: "WeeksAverage" },
  ];

  return (
    <div className="w-full overflow-scroll">
      <UITable>
        <TableHeader className="bg-blue-500 text-white">
          <TableRow className="hover:bg-blue-600">
            {headers.map((header, i) => (
              <TableHead
                className={cn(
                  "relative p-4 cursor-pointer text-white",
                  i > 0 && "text-right"
                )}
                key={JSON.stringify(header)}
                onClick={() => {
                  if (orderBy === header.key) {
                    setOrder({ orderBy, ascending: !ascending });
                  } else {
                    setOrder({ orderBy: header.key, ascending: true });
                  }
                }}
              >
                <span className="line-clamp-2">{header.title}</span>
                {header.key === orderBy && (
                  <div className="absolute bottom-1 right-1">
                    {!ascending ? " \u25BC" : " \u25B2"}
                  </div>
                )}
              </TableHead>
            ))}
          </TableRow>
        </TableHeader>

        <TableBody>
          {sortedUsers.map((user, i) => (
            <Row {...user} key={JSON.stringify(user)} />
          ))}
        </TableBody>
      </UITable>
    </div>
  );
};
