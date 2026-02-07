import { For, Show, createSignal } from "solid-js";
import { $servers } from "../stores/servers";
import { $expand } from "../stores/expand";
import { BiRegularRefresh } from "solid-icons/bi";
import { $notifications, NotificationType } from "../stores/notifications";
import { $busy } from "../stores/busy";
import { DropDownContainer } from "../components/Dropdown";
import { EmptyFallback } from "../components/EmptyFallback";
import { IoChevronCollapse, IoExpand } from "solid-icons/io";
import { FaSolidHashtag, FaSolidPen, FaSolidTrash } from "solid-icons/fa";
import { TiStarOutline, TiStarFullOutline } from "solid-icons/ti";
import { $tabs } from "../stores/tabs";
import { ServerWithReflection } from "../../bindings/grpc-gui";
import { ToggleFavoriteServer, DeleteServer } from "../../bindings/grpc-gui/app";
import { ContextMenu } from "@kobalte/core/context-menu";
import { VsRefresh } from "solid-icons/vs";
import { useNavigate } from "@solidjs/router";
import { ServerModal } from "./ServerModal";

export const WorkspaceServicesMenu = () => {
	const navigate = useNavigate();
	const { servers, isLoading } = $servers;
	const [searchQuery, setSearchQuery] = createSignal("");
	const [editingServer, setEditingServer] = createSignal<ServerWithReflection | null>(null);
	const [showAddModal, setShowAddModal] = createSignal(false);

	const filteredServers = () => {
		const query = searchQuery().toLowerCase().trim();
		if (!query) return servers();

		return servers().filter(server => {
			const serverName = server.server?.name?.toLowerCase() || "";

			if (serverName.includes(query)) return true;

			return server.reflection?.services?.some(service => {
				const serviceName = service.name?.toLowerCase() || "";
				if (serviceName.includes(query)) return true;

				return service.methods?.some(method => {
					const methodName = method.name?.toLowerCase() || "";
					return methodName.includes(query);
				}) || false;
			}) || false;
		});
	};

	const handleAddService = () => {
		setShowAddModal(true);
	};

	const handleToggleServerExpand = (key: string, value: boolean) => {
		$expand.setByKey(key, value);
	};

	const handleToggleServiceExpand = (key: string, value: boolean) => {
		$expand.setByKey(key, value);
	};

	const { busy } = $busy;
	const { expanded } = $expand;
	const {
		getServerExpandPersistentKey,
		getTabIdForMethod,
		getServiceExpandPersistentKey,
		getServerBusyPersistentKey,
	} = $servers;

	const handleOpenSendRequest = (server: ServerWithReflection, service: string, methodName: string) => {
		$tabs.openRequestTab(server.server?.id!, service, methodName);
	};

	const handleRefreshServer = async (serverId: number) => {
		$busy.lockByKey(getServerBusyPersistentKey(serverId), 1000);
		try {
			await $servers.refreshServerById(serverId);
		} catch (error) {
			$notifications.addNotification({
				type: NotificationType.ERROR,
				title: "Ошибка",
				message: "Не удалось обновить сервер",
			});
		} finally {
			await $busy.unlockByKey(getServerBusyPersistentKey(serverId));
		}
	};

	const handleToggleFavorite = async (e: MouseEvent, serverId: number) => {
		e.stopPropagation();
		$servers.toggleFavorite(serverId);
		try {
			await ToggleFavoriteServer(serverId);
		} catch (err) {
			$servers.toggleFavorite(serverId);
			$notifications.addNotification({
				type: NotificationType.ERROR,
				title: "Ошибка",
				message: "Не удалось изменить избранное",
			});
		}
	};

	const handleCollapseAllServers = () => {
		servers().forEach(server => {
			$expand.setByKey(getServerExpandPersistentKey(server.server?.id!), false);
			server.reflection?.services?.forEach(service => {
				$expand.setByKey(getServiceExpandPersistentKey(server.server?.id!, service.name), false);
			});
		});
	};

	const handleExpandAllServers = () => {
		servers().forEach(server => {
			$expand.setByKey(getServerExpandPersistentKey(server.server?.id!), true);
			server.reflection?.services?.forEach(service => {
				$expand.setByKey(getServiceExpandPersistentKey(server.server?.id!, service.name), true);
			});
		});
	};

	const handleDeleteServer = async (serverId: number) => {
		if (!confirm("Удалить сервер?")) return;
		
		try {
			await DeleteServer(serverId);
		} catch (err) {
			$notifications.addNotification({
				type: NotificationType.ERROR,
				title: "Ошибка",
				message: "Не удалось удалить сервер",
			});
			return;
		}
		
		$servers.removeServer(serverId);
		$notifications.addNotification({
			type: NotificationType.SUCCESS,
			title: "Успешно",
			message: "Сервер удален",
		});
	};

	const handleEditServer = (server: ServerWithReflection) => {
		setEditingServer(server);
	};

	const handleModalSuccess = async (serverId?: number, isEdit?: boolean) => {
		if (isEdit && serverId) {
			await $servers.refreshServerById(serverId);
		} else {
			await $servers.refreshServers();
		}
		setEditingServer(null);
		setShowAddModal(false);
	};

	return (
		<>
			<div class="flex justify-between items-center">
				<div class="font-extrabold text-sm text-base-content/70">Мои сервисы</div>

				<div class="flex gap-2 items-center">
					<Show when={!!servers().length}>
						<button
							title="Свернуть все"
							class="btn btn-xs btn-neutral btn-square"
							onClick={handleCollapseAllServers}>
							<IoChevronCollapse class="w-4 h-4" />
						</button>
						<button
							title="Развернуть все"
							class="btn btn-xs btn-neutral btn-square"
							onClick={handleExpandAllServers}>
							<IoExpand class="w-4 h-4" />
						</button>
					</Show>

					<button class="btn btn-xs btn-secondary" onClick={handleAddService}>
						Добавить
					</button>
				</div>
			</div>

			<Show when={servers().length > 0}>
				<div class="mt-2">
					<input
						type="text"
						class="input input-sm w-full"
						placeholder="Поиск по серверу, сервису или методу..."
						value={searchQuery()}
						onInput={e => setSearchQuery(e.currentTarget.value)}
					/>
				</div>
			</Show>

			<div class="mt-4 flex flex-col gap-1">
				<Show when={isLoading()}>
					<div class="flex items-center justify-center py-8">
						<span class="loading loading-spinner loading-md"></span>
					</div>
				</Show>

				<Show when={!isLoading()}>
					<For
						each={filteredServers()}
						fallback={
							<EmptyFallback
								message={
									searchQuery()
										? "Ничего не найдено"
										: "Пустота... Что есть пустота? Пустота... Это пустота..."
								}
							/>
						}>
						{server => {
							const serverId = server.server?.id!;
							const serverExpandKey = getServerExpandPersistentKey(serverId);
							const serverBusyKey = getServerBusyPersistentKey(serverId);

							return (
								<DropDownContainer
									open={expanded()[serverExpandKey]}
									title={server.server?.name!}
									prefix={
										<button
											class="hover:text-warning transition-colors cursor-default"
											onClick={e => handleToggleFavorite(e, serverId)}
											title={server.server?.favorite ? "Убрать из избранного" : "Добавить в избранное"}>
											{server.server?.favorite ? (
												<TiStarFullOutline class="w-3 h-3" />
											) : (
												<TiStarOutline class="w-3 h-3" />
											)}
										</button>
									}
									onOpenChange={v => handleToggleServerExpand(serverExpandKey, v)}
									headerWrapper={header => (
										<ContextMenu>
											<ContextMenu.Trigger class="w-full">{header}</ContextMenu.Trigger>
											<ContextMenu.Portal>
												<ContextMenu.Content class="context-menu">
													<ContextMenu.Item
														class="context-menu-item flex items-center gap-2"
														onSelect={() => handleRefreshServer(serverId)}>
														<VsRefresh class="w-3 h-3" />
														<span>Обновить рефлексию</span>
													</ContextMenu.Item>
													<ContextMenu.Item
														class="context-menu-item flex items-center gap-2"
														onSelect={() => handleEditServer(server)}>
														<FaSolidPen class="w-3 h-3" />
														<span>Редактировать</span>
													</ContextMenu.Item>
													<ContextMenu.Item
														class="context-menu-item context-menu-item-danger flex items-center gap-2"
														onSelect={() => handleDeleteServer(serverId)}>
														<FaSolidTrash class="w-3 h-3" />
														<span>Удалить</span>
													</ContextMenu.Item>
												</ContextMenu.Content>
											</ContextMenu.Portal>
										</ContextMenu>
									)}>
									<Show when={!server.error}>
										<For
											each={server.reflection?.services}
											fallback={<EmptyFallback message="А где сервисы?" />}>
											{service => {
												return (
													<DropDownContainer
														open={expanded()[getServiceExpandPersistentKey(serverId, service.name)]}
														title={service.name}
														onOpenChange={v =>
															handleToggleServiceExpand(
																getServiceExpandPersistentKey(serverId, service.name),
																v,
															)
														}>
														<div class="flex flex-col gap-0.5">
															<For
																each={service.methods}
																fallback={<EmptyFallback message="А где методы?" />}>
																{method => {
																	return (
																		<button
																			onClick={() => handleOpenSendRequest(server, service.name, method.name)}
																			class="w-full text-sm text-base-content/90 cursor-pointer flex gap-1 items-center px-2 justify-start hover:text-base-content transition-all duration-300 hover:bg-base-300 -ml-2">
																			<FaSolidHashtag class="w-2 h-2" />
																			<span class="truncate flex-1 text-left">{method.name}</span>
																		</button>
																	);
																}}
															</For>
														</div>
													</DropDownContainer>
												);
											}}
										</For>
									</Show>

									<Show when={server.error}>
										<div class="flex gap-1 items-center justify-between">
											<div title={server.error} class="truncate text-sm text-error/90">
												{server.error}
											</div>
											<button
												disabled={!!busy()[serverBusyKey]}
												onClick={() => handleRefreshServer(serverId)}
												class="btn btn-xs"
												title="Обновить, попробовать еще раз">
												<BiRegularRefresh
													class="w-4 h-4"
													classList={{ "animate-spin": !!busy()[serverBusyKey] }}
												/>
											</button>
										</div>
									</Show>
								</DropDownContainer>
							);
						}}
					</For>
				</Show>
			</div>

			<Show when={editingServer() || showAddModal()}>
				<ServerModal
					server={editingServer() || undefined}
					onClose={() => {
						setEditingServer(null);
						setShowAddModal(false);
					}}
					onSuccess={handleModalSuccess}
				/>
			</Show>
		</>
	);
};
