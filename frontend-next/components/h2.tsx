/** @jsx jsx */
import React from "react";
import { jsx, css } from "@emotion/react";

const weekTitle = css`
  font-family: Roboto, sans-serif;
  margin: 0px;
  padding: 20px;
  padding-top: 0px;
`;

export const H2: React.FC = ({ children }) => (
  <h2 css={weekTitle}>{children}</h2>
);
