import { createRoot, createSignal, onMount } from "solid-js";
import { GetServersWithReflection, GetServerWithReflection } from "../../bindings/grpc-gui/app";
import { ServerWithReflection } from "../../bindings/grpc-gui";
import { $notifications, NotificationType } from "./notifications";

const createServersStore = () => {
	const [servers, setServers] = createSignal<ServerWithReflection[]>([]);

	onMount(() => refreshServers());

	const refreshServers = async () => {
		try {
			const servers = await GetServersWithReflection();
			if (!servers) {
				throw new Error("GetServersWithReflection returned null");
			}
			setServers(servers);
		} catch (err) {
			return $notifications.addNotification({
				type: NotificationType.ERROR,
				title: "Ошибка",
				message: "Не удалось получить список сервисов",
			});
		}
	};

	const refreshServerById = async (serverId: number) => {
		try {
			const server = await GetServerWithReflection(serverId);
			setServers(s => s.map(v => (v.server?.id === serverId ? server! : v)));
		} catch (err) {
			return $notifications.addNotification({
				type: NotificationType.ERROR,
				title: "Ошибка",
				message: "Не удалось получить список сервисов",
			});
		}
	};

	const getServerExpandPersistentKey = (serverId: number) => {
		return `server-expand:${serverId}`;
	};

	const getServiceExpandPersistentKey = (serverId: number, serviceName: string) => {
		return `service-expand:${serverId}:${serviceName}`;
	};

	const getServerBusyPersistentKey = (serverId: number) => {
		return `server-busy:${serverId}`;
	};

	const getServiceBusyPersistentKey = (serverId: number, serviceName: string) => {
		return `service-busy:${serverId}:${serviceName}`;
	};

	const getTabIdForMethod = (serverId: number, serviceName: string, methodName: string) => {
		return `tab-${serverId}-${serviceName}-${methodName}`;
	};

	return {
		servers,
		refreshServers,
		getTabIdForMethod,
		refreshServerById,
		getServerExpandPersistentKey,
		getServiceExpandPersistentKey,
		getServerBusyPersistentKey,
		getServiceBusyPersistentKey,
	};
};

export const $servers = createRoot(createServersStore);
