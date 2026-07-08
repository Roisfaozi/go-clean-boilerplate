import { useState } from "react";

export function useAiChat() {
	const [messages, setMessages] = useState<any[]>([]);
	const [isTyping, setIsTyping] = useState(false);

	const sendMessage = async (content: string) => {
		setIsTyping(true);
		setMessages((prev) => [...prev, { role: "user", content }]);

		// Simulate AI response
		setTimeout(() => {
			setMessages((prev) => [
				...prev,
				{ role: "ai", content: "This is a placeholder response." },
			]);
			setIsTyping(false);
		}, 1000);
	};

	return { messages, sendMessage, isTyping };
}
