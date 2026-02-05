import { createRoot, createSignal } from "solid-js";
import { GetHistory } from "../../bindings/grpc-gui/app";
import { History } from "../../bindings/grpc-gui/internal/models/models";

const createHistoryStore = () => {
	const [history, setHistory] = createSignal<History[]>([]);
	const [isLoading, setIsLoading] = createSignal(false);

	const loadHistory = async () => {
		setIsLoading(true);
		try {
			const data = await GetHistory(0, 100); // 0 = все серверы, 100 = последние 100 запросов
			setHistory(data || []);
		} catch (error) {
			console.error("Failed to load history:", error);
		} finally {
			setIsLoading(false);
		}
	};

	const refresh = () => {
		loadHistory();
	};

	return {
		history,
		isLoading,
		loadHistory,
		refresh,
	};
};

export const $history = createRoot(createHistoryStore);
