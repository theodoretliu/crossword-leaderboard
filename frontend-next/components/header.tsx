import { jsx, css } from "@emotion/react";
import Link from "next/link";
import * as styles from "./header_styles";

export function Header() {
  return (
    <div css={styles.headerContainer}>
      <Link href="/" css={styles.headerTitle}>
        <h1>Mini Crossword Leaderboard</h1>
      </Link>

      <div css={styles.rightLinks}>
        <Link href="/previous_weeks" css={styles.headerLinks}>
          Previous Weeks
        </Link>

        <Link href="/settings" passHref css={styles.headerLinks}>
          Settings
        </Link>
      </div>
    </div>
  );
}
