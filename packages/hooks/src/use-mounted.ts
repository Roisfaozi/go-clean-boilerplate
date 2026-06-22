import { useSyncExternalStore } from "react";

const emptySubscribe = () => () => undefined;

/**
 * Hook to detect if the component has mounted on the client.
 * Useful for avoiding SSR hydration mismatches when using client-only APIs.
 */
export function useMounted() {
  const isMounted = useSyncExternalStore(
    emptySubscribe,
    () => true,
    () => false,
  );
  return isMounted;
}
