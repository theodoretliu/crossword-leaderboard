/** @jsx jsx */
import { jsx, css } from "@emotion/core";
import Link from "next/link";
import * as styles from "./header_styles";

export function Header() {
  return (
    <div css={styles.headerContainer}>
      <Link href="/" passHref>
        <a css={styles.headerTitle}>
          <h1>Mini Crossword Leaderboard</h1>
        </a>
      </Link>

      <div css={styles.rightLinks}>
        <Link href="/previous_weeks" passHref>
          <a css={styles.headerLinks}>Previous Weeks</a>
        </Link>

        <Link href="/settings" passHref>
          <a css={styles.headerLinks}>Settings</a>
        </Link>
      </div>
    </div>
  );
}
