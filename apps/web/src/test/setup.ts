import { cleanup } from "@testing-library/react";
import { afterEach } from "vitest";

afterEach(() => {
	cleanup();
});

if (typeof window !== "undefined") {
	Object.defineProperty(window, "matchMedia", {
		writable: true,
		value: (query: string): MediaQueryList =>
			({
				matches: false,
				media: query,
				onchange: null,
				addListener: () => {},
				removeListener: () => {},
				addEventListener: () => {},
				removeEventListener: () => {},
				dispatchEvent: () => false,
			}) as unknown as MediaQueryList,
	});

	class ResizeObserverShim {
		observe() {}
		unobserve() {}
		disconnect() {}
	}

	window.ResizeObserver ??= ResizeObserverShim;
	HTMLElement.prototype.scrollIntoView ??= () => {};
	HTMLElement.prototype.scrollTo ??= () => {};
}
