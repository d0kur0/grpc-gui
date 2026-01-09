import { For, Show } from "solid-js";
import { Viewport } from "../components/Viewport";
import { $servers } from "../stores/servers";
import "./Workspace.css";
import { useNavigate } from "@solidjs/router";
import { $expand } from "../stores/expand";
import { BiRegularRefresh } from "solid-icons/bi";
import { $notifications, NotificationType } from "../stores/notifications";
import { $busy } from "../stores/busy";
import { DropDownContainer } from "../components/Dropdown";
import { EmptyFallback } from "../components/EmptyFallback";
import { IoChevronCollapse, IoExpand } from "solid-icons/io";
import { FaSolidHashtag } from "solid-icons/fa";

export const Workspace = () => {
	return (
		<Viewport subtitle="workspace">
			<div class="workspace">
				<WorkspaceMenu />

				<div class="workspace__body">body</div>
			</div>
		</Viewport>
	);
};

const WorkspaceMenu = () => {
	const { servers } = $servers;
	const navigate = useNavigate();

	const handleAddService = () => {
		navigate("/add-service");
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
		getServiceExpandPersistentKey,
		getServerBusyPersistentKey,
		getServiceBusyPersistentKey,
	} = $servers;

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

	return (
		<div class="workspace__menu">
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

					<button class="btn btn-xs btn-primary" onClick={handleAddService}>
						Добавить
					</button>
				</div>
			</div>

			<div class="mt-4 flex flex-col gap-1">
				<For
					each={servers()}
					fallback={<EmptyFallback message="Пустота... Что есть пустота? Пустота... Это пустота..." />}>
					{server => {
						const serverId = server.server?.id!;
						const serverKey = getServerExpandPersistentKey(serverId);

						return (
							<DropDownContainer
								open={expanded()[serverKey]}
								title={server.server?.name!}
								onOpenChange={v => handleToggleServerExpand(serverKey, v)}>
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
															v
														)
													}>
													<For each={service.methods} fallback={<EmptyFallback message="А где методы?" />}>
														{method => {
															return (
																<button class="w-full text-sm text-base-content/90 cursor-pointer flex gap-1 items-center px-2 justify-start hover:text-base-content transition-all duration-300 hover:bg-base-300 -ml-2">
																	<FaSolidHashtag class="w-2 h-2" />
																	<span class="truncate">{method.name}</span>
																</button>
															);
														}}
													</For>
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
											disabled={!!busy()[serverKey]}
											onClick={() => handleRefreshServer(serverId)}
											class="btn btn-xs"
											title="Обновить, попробовать еще раз">
											<BiRegularRefresh class="w-4 h-4" classList={{ "animate-spin": !!busy()[serverKey] }} />
										</button>
									</div>
								</Show>
							</DropDownContainer>
						);
					}}
				</For>
			</div>
		</div>
	);
};
