import { Header } from "components/header";
import dayjs from "dayjs";
import utc from "dayjs/plugin/utc";
import Link from "next/link";

dayjs.extend(utc);

export default function PreviousWeeks() {
  let weekStarts = [];

  let currentWeek = dayjs.utc().startOf("week");

  for (let i = 0; i < 10; i++) {
    weekStarts.push(currentWeek);

    currentWeek = currentWeek.subtract(1, "week");
  }

  return (
    <div>
      <Header />
      <ul>
        {weekStarts.map((weekStart) => {
          return (
            <li>
              <Link
                href={`/week/${weekStart.year()}/${weekStart.month()}/${weekStart.date()}`}
              >
                {weekStart.toString()}
              </Link>
            </li>
          );
        })}
      </ul>
    </div>
  );
}
