import "../normalize.css";
import ApolloClient from "apollo-boost";
import { ApolloProvider } from "@apollo/react-hooks";
import { InMemoryCache } from "apollo-cache-inmemory";
import Head from "next/head";

export default function ({ Component, pageProps }) {
  const client = new ApolloClient({
    uri: "https://api-crossword.theodoretliu.com/graphql",
    cache: new InMemoryCache().restore(pageProps),
  });

  return (
    <ApolloProvider client={client}>
      <Head>
        <title>Mini Crossword Leaderboard</title>
      </Head>
      <Component />
    </ApolloProvider>
  );
}
