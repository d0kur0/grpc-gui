import { createRoot, createSignal, createEffect } from "solid-js";
import { createStore, produce } from "solid-js/store";
import { History } from "../../bindings/grpc-gui/internal/models/models";
import { DoGRPCRequest, SaveTabStates, GetTabStates, DeleteTabState } from "../../bindings/grpc-gui/app";
import { $history } from "./history";
import stripJsonComments from "strip-json-comments";

export enum TabType {
	REQUEST = "REQUEST",
}

export type KeyValuePair = {
	id: string;
	key: string;
	value: string;
};

export type SendRequestData = {
	serverId: number;
	serviceName: string;
	methodName: string;
	historyData?: History;
	activeTab: "body" | "metadata" | "context";
	requestBody: string;
	metadata: KeyValuePair[];
	contextValues: KeyValuePair[];
	response: string;
	responseTime: number;
};

export type TabData = {
	[TabType.REQUEST]: SendRequestData;
};

export type Tab<T extends TabType = TabType> = {
	id: string;
	name: string;
	type: T;
	data: TabData[T];
	isActive: boolean;
	temporary: boolean;
};

const createTabsStore = () => {
	const [tabs, setTabs] = createStore<Tab[]>([]);
	const [isLoaded, setIsLoaded] = createSignal(false);

	const addTab = <T extends TabType>(tab: Tab<T>) => {
		const existingIndex = tabs.findIndex(t => t.id === tab.id);

		if (existingIndex >= 0) {
			activateTab(tab.id);
			return;
		}

		if (tab.isActive) {
			setTabs(
				produce(tabs => {
					tabs.forEach(t => (t.isActive = false));
				}),
			);
		}

		setTabs(tabs.length, tab);
	};

	const removeTab = (id: string) => {
		const index = tabs.findIndex(tab => tab.id === id);
		if (index < 0) return;

		const wasActive = tabs[index].isActive;
		setTabs(
			produce(tabs => {
				tabs.splice(index, 1);
			}),
		);

		if (wasActive && tabs.length > 0) {
			const lastTab = tabs[tabs.length - 1];
			lastTab && activateTab(lastTab.id);
		}

		DeleteTabState(id).catch(err => console.error("Failed to delete tab state:", err));
	};

	const updateTabData = <T extends TabType>(id: string, data: Partial<TabData[T]>) => {
		const index = tabs.findIndex(t => t.id === id);
		if (index >= 0) {
			setTabs(
				index,
				"data",
				produce((existing: any) => {
					for (const key in data) {
						existing[key] = data[key];
					}
				}),
			);
		}
	};

	const getTabData = <T extends TabType>(id: string): TabData[T] | undefined => {
		return tabs.find(t => t.id === id)?.data as TabData[T] | undefined;
	};

	const addMetadataRow = (tabId: string) => {
		const tab = tabs.find(t => t.id === tabId);
		if (!tab || tab.type !== TabType.REQUEST) return;

		const data = tab.data as SendRequestData;
		updateTabData<TabType.REQUEST>(tabId, {
			metadata: [...data.metadata, { id: crypto.randomUUID(), key: "", value: "" }],
		});
	};

	const removeMetadataRow = (tabId: string, itemId: string) => {
		const tab = tabs.find(t => t.id === tabId);
		if (!tab || tab.type !== TabType.REQUEST) return;

		const data = tab.data as SendRequestData;
		updateTabData<TabType.REQUEST>(tabId, {
			metadata: data.metadata.filter(item => item.id !== itemId),
		});
	};

	const updateMetadataKey = (tabId: string, itemId: string, key: string) => {
		const tab = tabs.find(t => t.id === tabId);
		if (!tab || tab.type !== TabType.REQUEST) return;

		const data = tab.data as SendRequestData;
		updateTabData<TabType.REQUEST>(tabId, {
			metadata: data.metadata.map(item => (item.id === itemId ? { ...item, key } : item)),
		});
	};

	const updateMetadataValue = (tabId: string, itemId: string, value: string) => {
		const tab = tabs.find(t => t.id === tabId);
		if (!tab || tab.type !== TabType.REQUEST) return;

		const data = tab.data as SendRequestData;
		updateTabData<TabType.REQUEST>(tabId, {
			metadata: data.metadata.map(item => (item.id === itemId ? { ...item, value } : item)),
		});
	};

	const addContextRow = (tabId: string) => {
		const tab = tabs.find(t => t.id === tabId);
		if (!tab || tab.type !== TabType.REQUEST) return;

		const data = tab.data as SendRequestData;
		updateTabData<TabType.REQUEST>(tabId, {
			contextValues: [...data.contextValues, { id: crypto.randomUUID(), key: "", value: "" }],
		});
	};

	const removeContextRow = (tabId: string, itemId: string) => {
		const tab = tabs.find(t => t.id === tabId);
		if (!tab || tab.type !== TabType.REQUEST) return;

		const data = tab.data as SendRequestData;
		updateTabData<TabType.REQUEST>(tabId, {
			contextValues: data.contextValues.filter(item => item.id !== itemId),
		});
	};

	const updateContextKey = (tabId: string, itemId: string, key: string) => {
		const tab = tabs.find(t => t.id === tabId);
		if (!tab || tab.type !== TabType.REQUEST) return;

		const data = tab.data as SendRequestData;
		updateTabData<TabType.REQUEST>(tabId, {
			contextValues: data.contextValues.map(item => (item.id === itemId ? { ...item, key } : item)),
		});
	};

	const updateContextValue = (tabId: string, itemId: string, value: string) => {
		const tab = tabs.find(t => t.id === tabId);
		if (!tab || tab.type !== TabType.REQUEST) return;

		const data = tab.data as SendRequestData;
		updateTabData<TabType.REQUEST>(tabId, {
			contextValues: data.contextValues.map(item => (item.id === itemId ? { ...item, value } : item)),
		});
	};

	const activateTab = (id: string) => {
		setTabs(
			produce(tabs => {
				tabs.forEach(t => (t.isActive = t.id === id));
			}),
		);
	};

	const openRequestTab = (
		serverId: number,
		serviceName: string,
		methodName: string,
		historyData?: History,
	) => {
		const tabId = historyData
			? `history-${historyData.id}`
			: `request-${serverId}-${serviceName}-${methodName}`;

		const defaultMetadata = [{ id: crypto.randomUUID(), key: "", value: "" }];
		const defaultContextValues = [{ id: crypto.randomUUID(), key: "", value: "" }];

		let metadata = defaultMetadata;
		let contextValues = defaultContextValues;
		let requestBody = "{}";
		let response = "";
		let responseTime = 0;

		if (historyData) {
			if (historyData.request) {
				try {
					const parsed = JSON.parse(historyData.request);
					requestBody = JSON.stringify(parsed, null, 2);
				} catch {
					requestBody = historyData.request;
				}
			}

			if (historyData.requestHeaders) {
				try {
					const parsed = JSON.parse(historyData.requestHeaders);
					const metadataArray = Object.entries(parsed).map(([key, value]) => ({
						id: crypto.randomUUID(),
						key,
						value: String(value),
					}));
					if (metadataArray.length > 0) {
						metadata = metadataArray;
					}
				} catch (err) {
					console.error("Failed to parse request headers:", err);
				}
			}

			if (historyData.contextValues) {
				try {
					const parsed = JSON.parse(historyData.contextValues);
					const contextArray = Object.entries(parsed).map(([key, value]) => ({
						id: crypto.randomUUID(),
						key,
						value: String(value),
					}));
					if (contextArray.length > 0) {
						contextValues = contextArray;
					}
				} catch (err) {
					console.error("Failed to parse context values:", err);
				}
			}

			if (historyData.response) {
				try {
					const parsed = JSON.parse(historyData.response);
					const prettyJson = JSON.stringify(parsed, null, 2);
					const startTimeStr = new Date(historyData.createdAt).toLocaleString("ru-RU");
					response = `// Время начала выполнения запроса: ${startTimeStr}\n// Время выполнения запроса: ${historyData.executionTime}ms\n// Код ответа: ${historyData.statusCode}\n\n${prettyJson}`;
				} catch {
					const startTimeStr = new Date(historyData.createdAt).toLocaleString("ru-RU");
					response = `// Время начала выполнения запроса: ${startTimeStr}\n// Время выполнения запроса: ${historyData.executionTime}ms\n// Код ответа: ${historyData.statusCode}\n\n${historyData.response}`;
				}
			}

			responseTime = historyData.executionTime || 0;
		}

		addTab<TabType.REQUEST>({
			id: tabId,
			name: historyData ? `${serviceName}.${methodName} (История)` : `${serviceName}.${methodName}`,
			type: TabType.REQUEST,
			data: {
				serverId,
				serviceName,
				methodName,
				historyData,
				activeTab: "body",
				requestBody,
				metadata,
				contextValues,
				response,
				responseTime,
			},
			isActive: true,
			temporary: false,
		});
	};

	const saveTabs = async () => {
		if (!isLoaded()) return;

		try {
			const tabsData = tabs.map((tab, index) => ({
				tabId: tab.id,
				name: tab.name,
				component: tab.type,
				props: JSON.stringify({}),
				state: JSON.stringify(tab.data),
				isActive: tab.isActive,
				order: index,
			}));

			console.log("Saving tabs:", tabsData.length);
			await SaveTabStates(tabsData);
			console.log("Tabs saved successfully");
		} catch (err) {
			console.error("Failed to save tabs:", err);
		}
	};

	const loadTabs = async () => {
		try {
			const savedTabs = await GetTabStates();

			if (savedTabs && savedTabs.length > 0) {
				const restoredTabs: Tab[] = savedTabs.map(saved => ({
					id: saved.tabId,
					name: saved.name,
					type: saved.component as TabType,
					data: JSON.parse(saved.state || "{}"),
					isActive: saved.isActive,
					temporary: false,
				}));

				setTabs(restoredTabs);
			}
		} catch (err) {
			console.error("Failed to load tabs:", err);
		} finally {
			setIsLoaded(true);
		}
	};

	let saveTimeout: number | undefined;
	createEffect(() => {
		JSON.stringify(tabs);
		if (!isLoaded()) return;

		if (saveTimeout) clearTimeout(saveTimeout);
		saveTimeout = setTimeout(() => {
			saveTabs();
		}, 1000) as unknown as number;
	});

	const sendRequest = async (tabId: string, serverAddress: string) => {
		const tab = tabs.find(t => t.id === tabId);
		if (!tab || tab.type !== TabType.REQUEST) return;

		const data = tab.data as SendRequestData;

		try {
			updateTabData<TabType.REQUEST>(tabId, { response: "", responseTime: 0 });

			const startTime = new Date();
			const startTimeStr = startTime.toLocaleString("ru-RU");

			const metadataObj: { [key: string]: string } = {};
			data.metadata.forEach(item => {
				if (item.key.trim()) {
					metadataObj[item.key] = item.value;
				}
			});

			const contextObj: { [key: string]: string } = {};
			data.contextValues.forEach(item => {
				if (item.key.trim()) {
					contextObj[item.key] = item.value;
				}
			});

			const cleanedRequestBody = stripJsonComments(data.requestBody);

			const [responseData, time] = await DoGRPCRequest(
				data.serverId,
				serverAddress,
				data.serviceName,
				data.methodName,
				cleanedRequestBody,
				Object.keys(metadataObj).length > 0 ? metadataObj : null,
				Object.keys(contextObj).length > 0 ? contextObj : null,
			);

			let formattedResponse = "";
			try {
				const parsed = JSON.parse(responseData);
				const prettyJson = JSON.stringify(parsed, null, 2);
				formattedResponse = `// Время начала выполнения запроса: ${startTimeStr}\n// Время выполнения запроса: ${time}ms\n// Код ответа: 0 (OK)\n\n${prettyJson}`;
			} catch {
				formattedResponse = `// Время начала выполнения запроса: ${startTimeStr}\n// Время выполнения запроса: ${time}ms\n// Код ответа: 0 (OK)\n\n${responseData}`;
			}

			updateTabData<TabType.REQUEST>(tabId, {
				response: formattedResponse,
				responseTime: time,
			});

			$history.refresh();
		} catch (err: any) {
			let errorMessage = err?.message || "Неизвестная ошибка";

			try {
				const parsed = JSON.parse(errorMessage);
				errorMessage = JSON.stringify(parsed, null, 2);
			} catch {}

			const errorResponse = `// Ошибка выполнения запроса\n// Код ответа: 1 (ERROR)\n\n${errorMessage}`;
			updateTabData<TabType.REQUEST>(tabId, { response: errorResponse });

			$history.refresh();
		}
	};

	const closeAllTabs = () => {
		setTabs([]);
	};

	const closeOtherTabs = (id: string) => {
		const tab = tabs.find(t => t.id === id);
		if (!tab) return;

		setTabs([{ ...tab, isActive: true }]);
	};

	return {
		tabs,
		setTabs,
		addTab,
		removeTab,
		updateTabData,
		getTabData,
		activateTab,
		openRequestTab,
		sendRequest,
		loadTabs,
		saveTabs,
		isLoaded,
		closeAllTabs,
		closeOtherTabs,
		addMetadataRow,
		removeMetadataRow,
		updateMetadataKey,
		updateMetadataValue,
		addContextRow,
		removeContextRow,
		updateContextKey,
		updateContextValue,
	};
};

export const $tabs = createRoot(createTabsStore);
