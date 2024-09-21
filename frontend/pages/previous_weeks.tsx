import { jsx } from "@emotion/react";
import { useMemo } from "react";
import { Header } from "@/components/header";
import dayjs from "dayjs";
import utc from "dayjs/plugin/utc";
import Link from "next/link";

dayjs.extend(utc);

export default function PreviousWeeks() {
  const weekStarts = useMemo(() => {
    let weekStarts = [];

    let currentWeek = dayjs.utc().startOf("day");

    while (currentWeek.day() !== 1) {
      currentWeek = currentWeek.subtract(1, "day");
    }

    for (let i = 0; i < 52; i++) {
      currentWeek = currentWeek.subtract(1, "week");
      weekStarts.push(currentWeek);
    }

    return weekStarts;
  }, []);

  return (
    <div>
      <Header />

      <h2 className="px-4 pb-4 text-lg font-semibold">Previous Weeks</h2>

      <div className="flex flex-col gap-1 px-4 pb-4">
        {weekStarts.map((weekStart) => (
          <Link
            key={weekStart.toString()}
            href={`/week/${weekStart.year()}/${weekStart.month() + 1}/${weekStart.date()}`}
          >
            {weekStart.format("MMMM D, YYYY")}
          </Link>
        ))}
      </div>
    </div>
  );
}
