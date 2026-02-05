import { createRoot, createSignal, JSX } from "solid-js";
import { History } from "../../bindings/grpc-gui/internal/models/models";

export enum TabComponent {
	REQUEST,
}

export type Tab = {
	id: string;
	name: string;
	component: TabComponent;
	componentProps: Record<string, unknown>;
	isActive: boolean;
	temporary: boolean;
};

const createTabsStore = () => {
	const [tabs, setTabs] = createSignal<Tab[]>([]);

	const addTab = (tab: Tab) => {
		setTabs(prev => {
			const hasSameTab = prev.find(t => t.id === tab.id);

			if (tab.isActive) {
				prev = prev.map(t => ({ ...t, isActive: false }));
			}

			if (hasSameTab) {
				return prev.map(t => (t.id === tab.id ? { ...t, ...tab, isActive: true } : t));
			}

			return [...prev, tab];
		});
	};

	const removeTab = (id: string) => {
		const isActive = tabs().find(tab => tab.id === id)?.isActive;
		setTabs(prev => prev.filter(tab => tab.id !== id));
		if (isActive) {
			const lastTab = tabs()[tabs().length - 1];
			lastTab && activateTab(lastTab.id);
		}
	};

	const updateTab = (id: string, tab: Partial<Tab>) => {
		setTabs(prev => prev.map(t => (t.id === id ? { ...t, ...tab } : t)));
	};

	const activateTab = (id: string) => {
		setTabs(prev => prev.map(t => ({ ...t, isActive: t.id === id })));
	};

	const openHistoryRequest = (history: History) => {
		addTab({
			id: `history-${history.id}`,
			name: `${history.service} ${history.method} (История)`,
			component: TabComponent.REQUEST,
			componentProps: {
				serverId: history.serverId,
				serviceName: history.service,
				methodName: history.method,
				historyData: history,
			},
			isActive: true,
			temporary: false,
		});
	};

	return {
		tabs,
		addTab,
		removeTab,
		updateTab,
		activateTab,
		openHistoryRequest,
	};
};

export const $tabs = createRoot(createTabsStore);
