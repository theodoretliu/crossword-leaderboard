import React from "react";
import useSWR from "swr";

import { API_URL } from "api";

import * as z from "zod";

const FeatureFlagType = z.object({
  Status: z.boolean(),
});

export function useFeatureFlag(featureFlag: string): {
  loading: boolean;
  status: boolean;
  error?: any;
} {
  const { data, error } = useSWR(featureFlag, async (flag: string) => {
    let res = await fetch(`${API_URL}/feature_flag?flag=${encodeURIComponent(flag)}`);
    let json = await res.json();

    return FeatureFlagType.parse(json);
  });

  if (error) {
    return { loading: false, status: false, error };
  }

  if (!data) {
    return { loading: true, status: false };
  }

  return { loading: false, status: data.Status };
}
