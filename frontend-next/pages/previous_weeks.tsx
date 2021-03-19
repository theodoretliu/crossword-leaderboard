/** @jsx jsx */
import { jsx } from "@emotion/core";
import { useMemo } from "react";
import { Header } from "components/header";
import dayjs from "dayjs";
import utc from "dayjs/plugin/utc";
import Link from "next/link";

import * as styles from "components/previous_weeks_styles";

dayjs.extend(utc);

export default function PreviousWeeks() {
  const weekStarts = useMemo(() => {
    let weekStarts = [];

    let currentWeek = dayjs.utc().startOf("week");

    for (let i = 0; i < 10; i++) {
      currentWeek = currentWeek.subtract(1, "week");
      weekStarts.push(currentWeek);
    }

    return weekStarts;
  }, []);

  return (
    <div>
      <Header />
      <h2 css={styles.h2}>Previous Weeks</h2>

      <div css={styles.linkContainer}>
        {weekStarts.map((weekStart) => (
          <Link
            href={`/week/${weekStart.year()}/${weekStart.month()}/${weekStart.date()}`}
            passHref
          >
            <a css={styles.link}>{weekStart.format("MMMM D, YYYY")}</a>
          </Link>
        ))}
      </div>
    </div>
  );
}
