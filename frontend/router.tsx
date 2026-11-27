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

const POSTHOG_KEY = import.meta.env.VITE_POSTHOG_KEY;

// Loaded on demand so that builds without a key don't download PostHog at all.
const posthogReady = POSTHOG_KEY
  ? import("posthog-js").then(({ default: posthog }) => {
      posthog.init(POSTHOG_KEY, {
        api_host: import.meta.env.VITE_POSTHOG_HOST || "https://us.i.posthog.com",
        person_profiles: "identified_only",
        // Page views are captured from router navigation below.
        capture_pageview: false,
        loaded: (posthog) => {
          if (import.meta.env.DEV) posthog.debug();
        },
      });
      return posthog;
    })
  : undefined;

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

if (posthogReady) {
  router.subscribe("onResolved", () => {
    void posthogReady.then((posthog) => posthog.capture("$pageview"));
  });
}

declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router;
  }
}
