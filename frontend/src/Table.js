/** @jsx jsx */
import { jsx } from "@emotion/core";

export function Table(props) {
  let { headers, rows } = props;

  return (
    <div>
      {headers.map((header) => (
        <div>{header}</div>
      ))}
      {rows.map((rows) => (
        <div>
          {rows.map((row) => (
            <div>{row}</div>
          ))}
        </div>
      ))}
    </div>
  );
}
