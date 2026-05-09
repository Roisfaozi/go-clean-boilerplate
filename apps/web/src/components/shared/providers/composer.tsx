"use client";

import React, { ReactNode } from "react";

interface ProviderComposerProps {
  providers: Array<React.JSXElementConstructor<React.PropsWithChildren<any>>>;
  children: ReactNode;
}

/**
 * ProviderComposer - Flattens nested providers into a single linear list.
 * Helps prevent the "wrapper hell" in root layouts.
 */
export function ProviderComposer({
  providers,
  children,
}: ProviderComposerProps) {
  return (
    <>
      {providers.reduceRight((acc, Provider) => {
        return <Provider>{acc}</Provider>;
      }, children)}
    </>
  );
}
