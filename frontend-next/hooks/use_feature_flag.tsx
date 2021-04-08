import React from "react";
import useSWR from "swr";

import { API_URL } from "api";

import * as s from "superstruct";

const FeatureFlagType = s.object({
  Status: s.boolean(),
});

export function useFeatureFlag(
  featureFlag: string
): { loading: boolean; status: boolean; error?: any } {
  const { data, error } = useSWR(featureFlag, async (flag: string) => {
    let res = await fetch(
      `${API_URL}/feature_flag?flag=${encodeURIComponent(flag)}`
    );
    let json = await res.json();

    s.assert(json, FeatureFlagType);

    return json;
  });

  if (error) {
    return { loading: false, status: false, error };
  }

  if (!data) {
    return { loading: true, status: false };
  }

  return { loading: false, status: data.Status };
}
