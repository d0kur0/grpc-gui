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

	const toggleFavorite = (serverId: number) => {
		setServers(s => {
			const updated = s.map(v => {
				if (v.server?.id === serverId) {
					return {
						...v,
						server: v.server ? { ...v.server, favorite: !v.server.favorite } : undefined,
					};
				}
				return v;
			});
			
			return updated.sort((a, b) => {
				const aFav = a.server?.favorite ? 1 : 0;
				const bFav = b.server?.favorite ? 1 : 0;
				if (aFav !== bFav) {
					return bFav - aFav;
				}
				const aTime = a.server?.createdAt ? new Date(a.server.createdAt).getTime() : 0;
				const bTime = b.server?.createdAt ? new Date(b.server.createdAt).getTime() : 0;
				return bTime - aTime;
			});
		});
	};

	return {
		servers,
		refreshServers,
		getTabIdForMethod,
		refreshServerById,
		toggleFavorite,
		getServerExpandPersistentKey,
		getServiceExpandPersistentKey,
		getServerBusyPersistentKey,
		getServiceBusyPersistentKey,
	};
};

export const $servers = createRoot(createServersStore);
