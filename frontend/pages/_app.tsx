import "@/global.css";
import "@/fonts.css";
import React from "react";
import Head from "next/head";
import { AppProps } from "next/app";

export default function App({ Component, pageProps }: AppProps) {
  return (
    <React.Fragment>
      <Head>
        <title>Mini Crossword Leaderboard</title>
      </Head>
      <Component {...pageProps} />
    </React.Fragment>
  );
}
