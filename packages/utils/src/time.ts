export type EpochMs = number & { readonly __brand: unique symbol };
export type BusinessDate = string & { readonly __brand: unique symbol };

/**
 * Formats epoch ms into localized string with mandatory explicit timeZone.
 */
export function formatEpochMs(
	epochMs: number | EpochMs,
	timeZone: string,
	options?: Intl.DateTimeFormatOptions,
): string {
	if (!timeZone) {
		throw new Error("timeZone parameter is required for date formatting");
	}
	const date = new Date(epochMs);
	return new Intl.DateTimeFormat("en-US", {
		...options,
		timeZone,
	}).format(date);
}

/**
 * Formats epoch ms for medical records, appending explicit zone label (WIB, WITA, WIT, etc.).
 */
export function formatMedicalStamp(
	epochMs: number | EpochMs,
	timeZone: string,
): string {
	if (!timeZone) {
		throw new Error(
			"timeZone parameter is required for medical stamp formatting",
		);
	}
	const dateStr = formatEpochMs(epochMs, timeZone, {
		year: "numeric",
		month: "2-digit",
		day: "2-digit",
		hour: "2-digit",
		minute: "2-digit",
		second: "2-digit",
		hour12: false,
	});

	let zoneLabel = timeZone;
	if (timeZone === "Asia/Jakarta") zoneLabel = "WIB";
	else if (timeZone === "Asia/Makassar") zoneLabel = "WITA";
	else if (timeZone === "Asia/Jayapura") zoneLabel = "WIT";

	return `${dateStr} ${zoneLabel}`;
}
