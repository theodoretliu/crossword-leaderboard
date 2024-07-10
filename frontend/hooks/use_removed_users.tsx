import React, { useEffect, useState } from "react";
import * as z from "zod";

const removedUsersSchema = z.array(z.number());

const removedUsersKey = "removedUsers";

export const useRemovedUsers = () => {
  const [removedUsers, setRemovedUsers] = useState<number[]>(() => {
    if (typeof window === "undefined") {
      return [];
    }

    let parsed = window.localStorage.getItem(removedUsersKey);

    if (!parsed) {
      return [];
    }

    const stored = JSON.parse(parsed);

    const validated = removedUsersSchema.safeParse(stored);

    if (validated.success) {
      return validated.data;
    }

    return [];
  });

  useEffect(() => {
    window.localStorage.setItem("removedUsers", JSON.stringify(removedUsers));
  }, [removedUsers]);

  return [removedUsers, setRemovedUsers] as const;
};
