import { createRoot, createSignal } from "solid-js";

type ExpandMap = Record<string, boolean>;

const LOCAL_STORAGE_KEY = "expand-persistence-values";

const getInitialValues = () => {
	const values = localStorage.getItem(LOCAL_STORAGE_KEY);
	if (values) {
		return JSON.parse(values);
	}
	return {};
};

const saveValues = (values: ExpandMap) => {
	localStorage.setItem(LOCAL_STORAGE_KEY, JSON.stringify(values));
};

const createExpandStore = () => {
	const [expanded, setExpanded] = createSignal<ExpandMap>(getInitialValues());

	const setByKey = (key: string, value: boolean) => {
		setExpanded(prev => ({ ...prev, [key]: value }));
		saveValues(expanded());
	};

	return {
		expanded,
		setByKey,
	};
};

export const $expand = createRoot(createExpandStore);
