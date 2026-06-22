/**
 * SSE (Server-Sent Events) Client
 * Connects to /api/v1/events with auto-reconnect
 */

export type SSEEventType =
	| "notification"
	| "activity"
	| "presence"
	| "system"
	| "user_update"
	| "org_update";

export interface SSEEvent {
	type: SSEEventType;
	data: Record<string, unknown>;
	timestamp: string;
}

type EventHandler = (event: SSEEvent) => void;

export class EventClient {
	private url: string;
	private eventSource: EventSource | null = null;
	private handlers: Map<string, Set<EventHandler>> = new Map();
	private globalHandlers: Set<EventHandler> = new Set();
	private reconnectAttempts = 0;
	private maxReconnectAttempts = 10;
	private baseDelay = 1000;
	private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
	private _connected = false;

	constructor(baseUrl = "/api/v1") {
		this.url = `${baseUrl}/events`;
	}

	get connected() {
		return this._connected;
	}

	connect(): void {
		if (this.eventSource) this.disconnect();

		const url = this.url;

		try {
			this.eventSource = new EventSource(url, { withCredentials: true });

			this.eventSource.onopen = () => {
				this._connected = true;
				this.reconnectAttempts = 0;
				this.emit({
					type: "system",
					data: { status: "connected" },
					timestamp: new Date().toISOString(),
				});
			};

			this.eventSource.onmessage = (e) => {
				try {
					const event: SSEEvent = JSON.parse(e.data);
					this.emit(event);
				} catch {
					// ignore malformed messages
				}
			};

			this.eventSource.onerror = () => {
				this._connected = false;
				this.eventSource?.close();
				this.eventSource = null;
				this.emit({
					type: "system",
					data: { status: "disconnected" },
					timestamp: new Date().toISOString(),
				});
				this.scheduleReconnect();
			};
		} catch {
			this.scheduleReconnect();
		}
	}

	disconnect(): void {
		if (this.reconnectTimer) {
			clearTimeout(this.reconnectTimer);
			this.reconnectTimer = null;
		}
		this.eventSource?.close();
		this.eventSource = null;
		this._connected = false;
		this.reconnectAttempts = 0;
	}

	subscribe(type: SSEEventType | "*", handler: EventHandler): () => void {
		if (type === "*") {
			this.globalHandlers.add(handler);
			return () => this.globalHandlers.delete(handler);
		}
		if (!this.handlers.has(type)) this.handlers.set(type, new Set());
		this.handlers.get(type)!.add(handler);
		return () => this.handlers.get(type)?.delete(handler);
	}

	private emit(event: SSEEvent): void {
		this.globalHandlers.forEach((h) => {
			h(event);
		});
		this.handlers.get(event.type)?.forEach((h) => {
			h(event);
		});
	}

	private scheduleReconnect(): void {
		if (this.reconnectAttempts >= this.maxReconnectAttempts) return;
		const delay = Math.min(this.baseDelay * 2 ** this.reconnectAttempts, 30000);
		this.reconnectAttempts++;
		this.reconnectTimer = setTimeout(() => this.connect(), delay);
	}
}

// Singleton
export const eventClient = new EventClient();
