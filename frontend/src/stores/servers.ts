import { createRoot, createSignal, onMount } from "solid-js";
import { GetServers } from "../../bindings/grpc-gui/app";
import { Server } from "../../bindings/grpc-gui/internal/models";

const createServersStore = () => {
	const [servers, setServers] = createSignal<Server[]>([]);

	onMount(async () => {
		const servers = await GetServers();
		if (!servers) {
			throw new Error("Failed to get servers");
		}

		setServers(servers);
	});

	return {
		servers,
	};
};

export const $servers = createRoot(createServersStore);
