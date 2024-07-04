import "@/global.css";
import "@/fonts.css";
import React from "react";
import Head from "next/head";
import { AppProps } from "next/app";
import { TooltipProvider } from "@/components/ui/tooltip";

export default function App({ Component, pageProps }: AppProps) {
  return (
    <React.Fragment>
      <Head>
        <title>Teddy's Mini Leaderboard</title>
      </Head>
      <TooltipProvider>
        <Component {...pageProps} />
      </TooltipProvider>
    </React.Fragment>
  );
}
