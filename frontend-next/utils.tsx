import { withApollo as _withApollo } from "next-apollo";
import ApolloClient, { InMemoryCache } from "apollo-boost";

const apolloClient = new ApolloClient({
  uri: process.env.NEXT_PUBLIC_GRAPHQL_URL || "http://localhost:8000/graphql",
  cache: new InMemoryCache(),
});

export const withApollo = _withApollo(apolloClient);

export const secondsToMinutes = (seconds: number) => {
  const minutes = Math.floor(seconds / 60);
  const secondsRemaining = seconds % 60;

  return `${minutes}:${secondsRemaining.toString().padStart(2, "0")}`;
};
