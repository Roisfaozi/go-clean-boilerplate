import { formatEpochMs, formatMedicalStamp } from "./time";

const epoch = 1788283800000;

try {
	formatEpochMs(epoch, "");
	throw new Error("expected throw");
} catch (e: any) {
	if (!e.message.includes("timeZone parameter is required")) {
		throw e;
	}
}

try {
	formatMedicalStamp(epoch, "");
	throw new Error("expected throw");
} catch (e: any) {
	if (!e.message.includes("timeZone parameter is required")) {
		throw e;
	}
}

const formattedWib = formatEpochMs(epoch, "Asia/Jakarta", {
	hour12: false,
	hour: "2-digit",
	minute: "2-digit",
});
if (!formattedWib.includes("00:30")) {
	throw new Error("WIB formatting failed");
}

const stampWib = formatMedicalStamp(epoch, "Asia/Jakarta");
if (!stampWib.includes("WIB")) {
	throw new Error("WIB stamp failed");
}

const stampWit = formatMedicalStamp(epoch, "Asia/Jayapura");
if (!stampWit.includes("WIT")) {
	throw new Error("WIT stamp failed");
}
