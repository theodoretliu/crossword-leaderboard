import "../normalize.css";
import React from "react";
import Head from "next/head";

export default function ({ Component, pageProps }) {
  return (
    <React.Fragment>
      <Head>
        <title>Mini Crossword Leaderboard</title>
      </Head>
      <Component {...pageProps} />
    </React.Fragment>
  );
}
