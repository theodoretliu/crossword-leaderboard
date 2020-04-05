/** @jsx jsx */
import { jsx, css } from "@emotion/core";
import { Link } from "react-router-dom";

export function Header() {
  return (
    <div
      css={css`
        display: flex;
        flex-direction: row;
        align-items: center;
      `}
    >
      <Link
        css={css`
          font-family: Roboto, sans-serif;
          text-decoration: none;
          color: black;
          padding: 20px;
        `}
        to="/"
      >
        <h1
          css={css`
            margin: 0px;
          `}
        >
          Mini Crossword Leaderboard
        </h1>
      </Link>

      <Link
        css={css`
          font-family: Roboto, sans-serif;
          text-decoration: none;
          color: black;
          padding: 20px;
          position: absolute;
          right: 0px;
        `}
        to="/settings"
      >
        Settings
      </Link>
    </div>
  );
}
