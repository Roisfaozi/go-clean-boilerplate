"use client";

import { useCallback, useEffect, useState } from "react";

export function useScroll(threshold: number) {
	const [scrolled, setScrolled] = useState(() => {
		if (typeof window !== "undefined") {
			return window.scrollY > threshold;
		}
		return false;
	});

	const onScroll = useCallback(() => {
		setScrolled(window.scrollY > threshold);
	}, [threshold]);

	useEffect(() => {
		window.addEventListener("scroll", onScroll, { passive: true });
		return () => window.removeEventListener("scroll", onScroll);
	}, [onScroll]);

	return scrolled;
}
