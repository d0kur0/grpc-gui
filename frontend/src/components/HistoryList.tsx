import { createSignal, Index, Show, onMount } from "solid-js";
import { DeleteHistoryItem } from "../../bindings/grpc-gui/app";
import { History } from "../../bindings/grpc-gui/internal/models/models";
import { $tabs } from "../stores/tabs";
import { $notifications, NotificationType } from "../stores/notifications";
import { $history } from "../stores/history";
import { EmptyFallback } from "./EmptyFallback";
import { IoTrash, IoTime } from "solid-icons/io";
import { FaSolidHashtag } from "solid-icons/fa";

export const HistoryList = () => {
	const { history, isLoading, loadHistory } = $history;
	const [searchQuery, setSearchQuery] = createSignal("");

	const filteredHistory = () => {
		const query = searchQuery().toLowerCase().trim();
		if (!query) return history();

		return history().filter(item => {
			const serviceName = item.service.toLowerCase();
			const methodName = item.method.toLowerCase();
			const fullName = `${serviceName} ${methodName}`;
			
			return serviceName.includes(query) || 
				   methodName.includes(query) || 
				   fullName.includes(query);
		});
	};

	onMount(() => {
		loadHistory();
	});

	const handleOpenRequest = (item: History) => {
		$tabs.openHistoryRequest(item);
	};

	const handleDeleteItem = async (e: MouseEvent, id: number) => {
		e.stopPropagation();
		try {
			await DeleteHistoryItem(id);
			await loadHistory();
			$notifications.addNotification({
				type: NotificationType.SUCCESS,
				title: "Удалено",
				message: "Запрос удален из истории",
			});
		} catch (error) {
			$notifications.addNotification({
				type: NotificationType.ERROR,
				title: "Ошибка",
				message: "Не удалось удалить запрос",
			});
		}
	};

	const formatDate = (dateStr: string) => {
		const date = new Date(dateStr);
		return date.toLocaleString('ru-RU', {
			day: '2-digit',
			month: '2-digit',
			year: 'numeric',
			hour: '2-digit',
			minute: '2-digit',
		});
	};

	return (
		<>
			<div class="flex justify-between items-center">
				<div class="font-extrabold text-sm text-base-content/70">История запросов</div>
				<button 
					class="btn btn-xs btn-neutral" 
					onClick={loadHistory}
					disabled={isLoading()}
				>
					Обновить
				</button>
			</div>

			<div class="mt-2">
				<input
					type="text"
					class="input input-sm w-full"
					placeholder="Поиск по сервису или методу..."
					value={searchQuery()}
					onInput={(e) => setSearchQuery(e.currentTarget.value)}
				/>
			</div>

			<div class="mt-4 flex flex-col gap-1">
				<Show when={!isLoading()} fallback={<div class="text-center text-base-content/50">Загрузка...</div>}>
					<Index
						each={filteredHistory()}
						fallback={
							<EmptyFallback 
								message={searchQuery() ? "Ничего не найдено" : "История пуста"} 
							/>
						}
					>
						{(item, index) => (
							<div
								class="flex items-center gap-2 p-2 hover:bg-base-300 rounded cursor-pointer transition-all duration-300 group"
								onClick={() => handleOpenRequest(item())}
							>
								<div class="flex-1 min-w-0">
									<div class="flex items-center gap-1 text-sm font-medium text-base-content truncate">
										<FaSolidHashtag class="w-2 h-2 flex-shrink-0" />
										<span class="truncate">{item().service} / {item().method}</span>
									</div>
									<div class="flex items-center gap-2 text-xs text-base-content/60 mt-0.5">
										<IoTime class="w-3 h-3 flex-shrink-0" />
										<span>{formatDate(item().createdAt)}</span>
										<span class="badge badge-xs badge-ghost">{item().executionTime}ms</span>
										<Show when={item().statusCode === 0}>
											<span class="badge badge-xs badge-success">OK</span>
										</Show>
										<Show when={item().statusCode !== 0}>
											<span class="badge badge-xs badge-error">ERROR</span>
										</Show>
									</div>
								</div>
								<button
									class="btn btn-xs btn-ghost btn-square opacity-0 group-hover:opacity-100 transition-opacity"
									onClick={(e) => handleDeleteItem(e, item().id)}
									title="Удалить"
								>
									<IoTrash class="w-4 h-4" />
								</button>
							</div>
						)}
					</Index>
				</Show>
			</div>
		</>
	);
};
