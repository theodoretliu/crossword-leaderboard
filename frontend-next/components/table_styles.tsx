import { css } from "@emotion/core";

export const table = css`
  width: 100%;
  height: auto;
  border-spacing: 0px;
`;

export const tableRow = css`
  td {
    padding: 20px;
    font-family: "Roboto", sans-serif;
    text-align: center;
  }
`;

export const th = css`
  position: relative;
  padding: 20px;
  font-family: "Roboto", sans-serif;
  font-weight: bold;
  color: white;
  background-color: #4d88f8;
  cursor: pointer;
`;

export const arrow = css`
  position: absolute;
  font-size: 12px;
  right: 5px;
  bottom: 5px;
`;

export const tbody = css`
  tr:nth-of-type(even) {
    background-color: #e0e0e0;
  }
`;
