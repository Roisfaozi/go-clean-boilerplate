"use client";

import { useSyncExternalStore } from "react";

const emptySubscribe = () => () => {};

export function useMounted() {
  const isMounted = useSyncExternalStore(
    emptySubscribe,
    () => true,
    () => false
  );
  return isMounted;
}
