import "../normalize.css";
import "../fonts.css";
import React from "react";
import Head from "next/head";

interface AppProps<T> {
  Component: React.FC<T>;
  pageProps: T;
}

export default function App<T>({ Component, pageProps }: AppProps<T>) {
  return (
    <React.Fragment>
      <Head>
        <title>Mini Crossword Leaderboard</title>
      </Head>
      <Component {...pageProps} />
    </React.Fragment>
  );
}
