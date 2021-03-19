import { css } from "@emotion/core";

export const headerContainer = css`
  display: flex;
  flex-direction: row;
  align-items: center;
`;

export const headerTitle = css`
  font-family: Roboto, sans-serif;
  text-decoration: none;
  color: black;
  padding: 20px;
  cursor: pointer;

  h1 {
    margin: 0px;
  }
`;

export const rightLinks = css`
  position: absolute;
  right: 10px;
`;

export const headerLinks = css`
  font-family: Roboto, sans-serif;
  text-decoration: none;
  color: black;
  padding: 0px 10px;
  cursor: pointer;
`;
