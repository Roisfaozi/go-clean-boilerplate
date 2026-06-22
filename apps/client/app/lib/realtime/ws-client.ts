/**
 * WebSocket Client
 * Connects to /api/v1/ws with auto-reconnect, ping/pong, and typed messaging
 */
import { authApi } from "../api/auth";

export type WSMessageType =
	| "presence_update"
	| "presence_join"
	| "presence_leave"
	| "chat"
	| "typing"
	| "ping"
	| "pong"
	| "metrics_update";

export interface WSMessage {
	type: WSMessageType;
	data: Record<string, unknown>;
	timestamp?: string;
}

type WSHandler = (message: WSMessage) => void;

export class WebSocketClient {
	private url: string;
	private ws: WebSocket | null = null;
	private handlers: Map<string, Set<WSHandler>> = new Map();
	private globalHandlers: Set<WSHandler> = new Set();
	private reconnectAttempts = 0;
	private maxReconnectAttempts = 10;
	private baseDelay = 1000;
	private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
	private pingInterval: ReturnType<typeof setInterval> | null = null;
	private _connected = false;

	constructor(baseUrl = "/api/v1") {
		if (typeof window !== "undefined") {
			const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
			const host = window.location.host;
			this.url = `${protocol}//${host}${baseUrl}/ws`;
		} else {
			this.url = "";
		}
	}

	get connected() {
		return this._connected;
	}

	async connect(): Promise<void> {
		if (typeof window === "undefined") return;
		if (this.reconnectTimer) {
			clearTimeout(this.reconnectTimer);
			this.reconnectTimer = null;
		}
		if (this.ws) this.disconnect();

		let ticket = "";
		try {
			const response = await authApi.getWsTicket();
			ticket = response.ticket;
		} catch (error) {
			console.error("Failed to fetch WS ticket:", error);
			this.scheduleReconnect();
			return;
		}

		const url = ticket
			? `${this.url}?ticket=${encodeURIComponent(ticket)}`
			: this.url;

		try {
			this.ws = new WebSocket(url);

			this.ws.onopen = () => {
				this._connected = true;
				this.reconnectAttempts = 0;
				this.startPing();
				this.emit({ type: "presence_update", data: { status: "connected" } });
			};

			this.ws.onmessage = (e) => {
				try {
					const msg: WSMessage = JSON.parse(e.data);
					if (msg.type === "pong") return; // swallow pong
					this.emit(msg);
				} catch {
					// ignore
				}
			};

			this.ws.onclose = () => {
				this._connected = false;
				this.stopPing();
				this.emit({
					type: "presence_update",
					data: { status: "disconnected" },
				});
				this.scheduleReconnect();
			};

			this.ws.onerror = () => {
				this.ws?.close();
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
		this.stopPing();
		this.ws?.close();
		this.ws = null;
		this._connected = false;
		this.reconnectAttempts = 0;
	}

	send(message: WSMessage): void {
		if (this.ws?.readyState === WebSocket.OPEN) {
			this.ws.send(JSON.stringify(message));
		}
	}

	subscribe(type: WSMessageType | "*", handler: WSHandler): () => void {
		if (type === "*") {
			this.globalHandlers.add(handler);
			return () => this.globalHandlers.delete(handler);
		}
		if (!this.handlers.has(type)) this.handlers.set(type, new Set());
		this.handlers.get(type)!.add(handler);
		return () => this.handlers.get(type)?.delete(handler);
	}

	private emit(message: WSMessage): void {
		this.globalHandlers.forEach((h) => {
			h(message);
		});
		this.handlers.get(message.type)?.forEach((h) => {
			h(message);
		});
	}

	private startPing(): void {
		this.pingInterval = setInterval(() => {
			this.send({ type: "ping", data: {} });
		}, 30000);
	}

	private stopPing(): void {
		if (this.pingInterval) {
			clearInterval(this.pingInterval);
			this.pingInterval = null;
		}
	}

	private scheduleReconnect(): void {
		if (this.reconnectAttempts >= this.maxReconnectAttempts) return;
		const delay = Math.min(this.baseDelay * 2 ** this.reconnectAttempts, 30000);
		this.reconnectAttempts++;
		this.reconnectTimer = setTimeout(() => this.connect(), delay);
	}
}

// Singleton
export const wsClient = new WebSocketClient();
