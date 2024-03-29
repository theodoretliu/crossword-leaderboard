import { jsx, css } from "@emotion/react";
import Link from "next/link";
import * as styles from "./header_styles";

export function Header() {
  return (
    <div css={styles.headerContainer}>
      <Link href="/" passHref legacyBehavior>
        <a css={styles.headerTitle}>
          <h1>Mini Crossword Leaderboard</h1>
        </a>
      </Link>

      <div css={styles.rightLinks}>
        <Link href="/previous_weeks" passHref legacyBehavior>
          <a css={styles.headerLinks}>Previous Weeks</a>
        </Link>

        <Link href="/settings" passHref legacyBehavior>
          <a css={styles.headerLinks}>Settings</a>
        </Link>
      </div>
    </div>
  );
}
