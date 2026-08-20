import { useEffect } from "react";
import { eventClient } from "@/lib/realtime";
import { wsClient } from "@/lib/realtime";
import {
	useNotificationStore,
	usePresenceStore,
	useActivityStore,
	useConnectionStore,
	useMetricsStore,
} from "@/stores/realtime-store";
import { toast } from "@casbin/ui";

/**
 * Hook that initializes realtime connections (SSE + WebSocket)
 * and wires events to stores. Mount once at app root level.
 */
export function useRealtimeInit() {
	const addNotification = useNotificationStore((s) => s.addNotification);
	const addActivity = useActivityStore((s) => s.addActivity);
	const { addUser, removeUser, updateUser } = usePresenceStore();
	const { setSseConnected, setWsConnected } = useConnectionStore();
	const updateMetrics = useMetricsStore((s) => s.updateMetrics);

	useEffect(() => {
		// ── SSE subscriptions ──
		const unsubSystem = eventClient.subscribe("system", (e) => {
			const status = e.data.status as string;
			setSseConnected(status === "connected");
		});

		const unsubNotification = eventClient.subscribe("notification", (e) => {
			const { title, description, type } = e.data as {
				title: string;
				description?: string;
				type?: "info" | "success" | "warning" | "danger" | "system";
			};
			addNotification({ title, description, type: type || "info" });
			// Also show a toast
			toast({
				title,
				description,
				variant: type === "danger" ? "destructive" : "default",
			});
		});

		const unsubActivity = eventClient.subscribe("activity", (e) => {
			const { user, action, target, type } = e.data as {
				user: string;
				action: string;
				target?: string;
				type?: "info" | "success" | "warning" | "danger";
			};
			addActivity({ user, action, target, type: type || "info" });
		});

		eventClient.connect();

		// ── WebSocket subscriptions ──
		const unsubPresenceJoin = wsClient.subscribe("presence_join", (msg) => {
			const user = msg.data as { id: string; name: string; avatar?: string };
			addUser({ ...user, status: "online" });
		});

		const unsubPresenceLeave = wsClient.subscribe("presence_leave", (msg) => {
			removeUser(msg.data.id as string);
		});

		const unsubPresenceUpdate = wsClient.subscribe("presence_update", (msg) => {
			if (msg.data.status === "connected") {
				setWsConnected(true);
			} else if (msg.data.status === "disconnected") {
				setWsConnected(false);
			} else if (msg.data.id) {
				updateUser(
					msg.data.id as string,
					msg.data as Partial<{ status: "online" | "away" | "busy" }>,
				);
			}
		});

		const unsubMetrics = wsClient.subscribe("metrics_update", (msg) => {
			updateMetrics(msg.data as any);
		});

		wsClient.connect();

		return () => {
			unsubSystem();
			unsubNotification();
			unsubActivity();
			unsubPresenceJoin();
			unsubPresenceLeave();
			unsubPresenceUpdate();
			unsubMetrics();
			eventClient.disconnect();
			wsClient.disconnect();
		};
	}, [
		addActivity,
		addNotification,
		addUser,
		removeUser,
		setSseConnected,
		setWsConnected,
		updateMetrics,
		updateUser,
	]); // eslint-disable-line react-hooks/exhaustive-deps
}
