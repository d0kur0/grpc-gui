import { For } from "solid-js";
import { Viewport } from "../components/Viewport";
import { $servers } from "../stores/servers";
import "./Workspace.css";

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

	return (
		<div class="workspace__menu">
			<div class="flex justify-between items-center">
				<div class="font-extrabold text-sm text-base-content/70">Мои сервисы</div>
				<button class="btn btn-xs btn-primary">Добавить сервис</button>
			</div>

			<div class="mt-6">
				<For each={servers()} fallback={<EmptyServersFallback />}>
					{server => (
						<div class="flex items-center gap-2">
							<div class="w-10 h-10 bg-base-content/10 rounded-full"></div>
							<div class="text-sm text-base-content/70">{server.name}</div>
						</div>
					)}
				</For>
			</div>
		</div>
	);
};

const EmptyServersFallback = () => {
	return (
		<div class="flex items-center justify-center text-base-content/50	text-sm border-3 border-base-content/10 rounded-md p-4 border-dashed">
			Пустота... Что есть пустота? Пустота... Это пустота...
		</div>
	);
};
