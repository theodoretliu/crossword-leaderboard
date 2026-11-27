import React from "react";
import { createRootRoute, createRoute, createRouter, Outlet } from "@tanstack/react-router";
import { TooltipProvider } from "@/components/ui/tooltip";
import { Index } from "@/routes/index";
import { Leaderboard } from "@/routes/leaderboard";
import { PreviousWeeks } from "@/routes/previous_weeks";
import { Settings } from "@/routes/settings";
import { User } from "@/routes/user";
import { Week } from "@/routes/week";

export const DEFAULT_TITLE = "Teddy's Mini Leaderboard";

function RootLayout() {
  return (
    <TooltipProvider>
      <Outlet />
    </TooltipProvider>
  );
}

const rootRoute = createRootRoute({ component: RootLayout });

const routeTree = rootRoute.addChildren([
  createRoute({ getParentRoute: () => rootRoute, path: "/", component: Index }),
  createRoute({ getParentRoute: () => rootRoute, path: "/leaderboard", component: Leaderboard }),
  createRoute({
    getParentRoute: () => rootRoute,
    path: "/previous_weeks",
    component: PreviousWeeks,
  }),
  createRoute({ getParentRoute: () => rootRoute, path: "/settings", component: Settings }),
  createRoute({ getParentRoute: () => rootRoute, path: "/users/$userId", component: User }),
  createRoute({
    getParentRoute: () => rootRoute,
    path: "/week/$year/$month/$day",
    component: Week,
  }),
]);

export const router = createRouter({ routeTree, defaultPreload: "intent" });

declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router;
  }
}
