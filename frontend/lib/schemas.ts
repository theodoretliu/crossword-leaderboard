import * as z from "zod";

export const UserType = z.object({
  UserId: z.number(),
  Username: z.string(),
  WeeksTimes: z.array(z.number()),
  WeeksAverage: z.number(),
  Elo: z.number(),
  Qualified: z.boolean(),
});

export const ResponseType = z.object({
  Users: z.array(UserType),
  DaysOfTheWeek: z.array(z.string()),
});
