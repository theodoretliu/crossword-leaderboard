import { withApollo as _withApollo } from "next-apollo";
import ApolloClient, { InMemoryCache } from "apollo-boost";

const apolloClient = new ApolloClient({
  uri: process.env.NEXT_PUBLIC_API_URL || "http://localhost:8000/graphql",
  cache: new InMemoryCache(),
});

export const withApollo = _withApollo(apolloClient);
