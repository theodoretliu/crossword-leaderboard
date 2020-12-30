/** @jsx jsx */
import { jsx, css } from "@emotion/core";
import Link from "next/link";

export function Header() {
  return (
    <div
      css={css`
        display: flex;
        flex-direction: row;
        align-items: center;
      `}
    >
      <Link href="/" passHref>
        <a
          css={css`
            font-family: Roboto, sans-serif;
            text-decoration: none;
            color: black;
            padding: 20px;
            cursor: pointer;
          `}
        >
          <h1
            css={css`
              margin: 0px;
            `}
          >
            Mini Crossword Leaderboard
          </h1>
        </a>
      </Link>

      <Link href="/settings" passHref>
        <a
          css={css`
            font-family: Roboto, sans-serif;
            text-decoration: none;
            color: black;
            padding: 20px;
            position: absolute;
            right: 0px;
            cursor: pointer;
          `}
        >
          Settings
        </a>
      </Link>
    </div>
  );
}
