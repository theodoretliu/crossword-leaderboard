import React from "react";
import ReactDOM from "react-dom";
import App from "./App";
import { Settings } from "./Settings";
import { ApolloProvider } from "@apollo/react-hooks";
import { BrowserRouter as Router, Switch, Route } from "react-router-dom";
import ApolloClient from "apollo-boost";
import "./normalize.css";

const client = new ApolloClient({
  uri: process.env.REACT_APP_GRAPHQL_URL,
});

ReactDOM.render(
  <React.StrictMode>
    <ApolloProvider client={client}>
      <Router>
        <Switch>
          <Route exact path="/">
            <App />
          </Route>

          <Route path="/settings">
            <Settings />
          </Route>
        </Switch>
      </Router>
    </ApolloProvider>
  </React.StrictMode>,
  document.getElementById("root")
);
