import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

/**
 * Merges class names using clsx and tailwind-merge.
 * This is the standard utility for conditional Tailwind classes.
 */
export function cn(...inputs: ClassValue[]) {
	return twMerge(clsx(inputs));
}

/**
 * Formats a number into a human-readable string with units (K, M, G, etc.).
 */
export function nFormatter(num: number, digits?: number) {
	if (!num) return "0";
	const lookup = [
		{ value: 1, symbol: "" },
		{ value: 1e3, symbol: "K" },
		{ value: 1e6, symbol: "M" },
		{ value: 1e9, symbol: "G" },
		{ value: 1e12, symbol: "T" },
		{ value: 1e15, symbol: "P" },
		{ value: 1e18, symbol: "E" },
	];
	const rx = /\.0+$|(\.[0-9]*[1-9])0+$/;
	const item = lookup
		.slice()
		.reverse()
		.find(function (item) {
			return num >= item.value;
		});
	return item
		? (num / item.value).toFixed(digits || 1).replace(rx, "$1") + item.symbol
		: "0";
}

/**
 * Checks if a filename contains spaces.
 */
export function hasFileNameSpaces(fileName: string) {
	return /\s/.test(fileName);
}

/**
 * Formats a date string or timestamp into a long date format (e.g., January 1, 2024).
 */
export function formatDate(input: string | number): string {
	const date = new Date(input);
	return date.toLocaleDateString("en-US", {
		month: "long",
		day: "numeric",
		year: "numeric",
	});
}
