import React from "react";
import { createRootRoute, createRoute, createRouter, Outlet } from "@tanstack/react-router";
import posthog from "posthog-js";
import { PostHogProvider } from "posthog-js/react";
import { TooltipProvider } from "@/components/ui/tooltip";
import { Index } from "@/routes/index";
import { Leaderboard } from "@/routes/leaderboard";
import { PreviousWeeks } from "@/routes/previous_weeks";
import { Settings } from "@/routes/settings";
import { User } from "@/routes/user";
import { Week } from "@/routes/week";

export const DEFAULT_TITLE = "Teddy's Mini Leaderboard";

const POSTHOG_KEY = import.meta.env.VITE_POSTHOG_KEY;

if (POSTHOG_KEY) {
  posthog.init(POSTHOG_KEY, {
    api_host: import.meta.env.VITE_POSTHOG_HOST || "https://us.i.posthog.com",
    person_profiles: "identified_only",
    // Page views are captured from router navigation below.
    capture_pageview: false,
    loaded: (posthog) => {
      if (import.meta.env.DEV) posthog.debug();
    },
  });
}

function RootLayout() {
  return (
    <PostHogProvider client={posthog}>
      <TooltipProvider>
        <Outlet />
      </TooltipProvider>
    </PostHogProvider>
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

if (POSTHOG_KEY) {
  router.subscribe("onResolved", () => posthog.capture("$pageview"));
}

declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router;
  }
}
