import { createEffect, createSignal, For, JSX, Show } from "solid-js";
import "./Viewport.css";
import { Window } from "@wailsio/runtime";
import { RiDevelopmentTerminalBoxFill } from "solid-icons/ri";
import { FaSolidArrowRightLong, FaSolidCircleInfo } from "solid-icons/fa";
import { useNavigate, useLocation } from "@solidjs/router";
import { $notifications, NotificationType } from "../stores/notifications";
import { AiTwotoneCheckCircle } from "solid-icons/ai";
import { BiRegularErrorCircle } from "solid-icons/bi";
import { IoWarning, IoTime, IoHome } from "solid-icons/io";

type ViewportProps = {
	subtitle?: string;
	children: JSX.Element;
};

export const Viewport = (props: ViewportProps) => {
	const [isMaximised, setIsMaximised] = createSignal(false);

	const navigate = useNavigate();
	const location = useLocation();

	createEffect(async () => {
		const isMaximised = await Window.IsMaximised();
		setIsMaximised(isMaximised);
	});

	const handleWindowMinimize = () => {
		Window.Minimise();
	};

	const handleWindowMaximize = async () => {
		isMaximised() ? Window.UnMaximise() : Window.Maximise();
		setIsMaximised(!isMaximised());
	};

	const handleWindowClose = () => {
		Window.Close();
	};

	return (
		<div class="viewport select-none">
			<div class="viewport__controls">
				<div class="viewport__title">
					<button class="cursor-pointer mr-1" onClick={() => navigate("/")}>
						grpc-gui
					</button>

					<For each={[
						{ path: "/", label: "Рабочая область" },
						{ path: "/history", label: "История запросов" }
					]}>
						{item => (
							<button
								class="btn transition-all duration-300 ease-in-out btn-xs hover:btn-accent hover:btn-soft"
								classList={{ "btn-accent btn-soft": location.pathname === item.path }}
								onClick={() => navigate(item.path)}>
								{item.label}
							</button>
						)}
					</For>
				</div>

				<div class="viewport__app-controls">

					<button title="Открыть DevTools" class="btn btn-xs" onClick={Window.OpenDevTools}>
						<RiDevelopmentTerminalBoxFill class="w-4 h-4" />
					</button>
				</div>

				<div class="viewport__window-controls">
					<button
						class="viewport__window-control viewport__window-control-minimize"
						onClick={handleWindowMinimize}></button>
					<button
						class="viewport__window-control viewport__window-control-maximize"
						onClick={handleWindowMaximize}></button>
					<button
						class="viewport__window-control viewport__window-control-close"
						onClick={handleWindowClose}></button>
				</div>
			</div>
			<div class="viewport__body">{props.children}</div>

			<Notifications />
		</div>
	);
};

const notificationClasses = {
	[NotificationType.SUCCESS]: "alert-success",
	[NotificationType.ERROR]: "alert-error",
	[NotificationType.WARNING]: "alert-warning",
	[NotificationType.INFO]: "alert-info",
};

const notificationIcons = {
	[NotificationType.SUCCESS]: AiTwotoneCheckCircle,
	[NotificationType.ERROR]: BiRegularErrorCircle,
	[NotificationType.WARNING]: IoWarning,
	[NotificationType.INFO]: FaSolidCircleInfo,
};

const Notifications = () => {
	const { notifications } = $notifications;

	return (
		<div class="absolute bottom-5 right-5 flex flex-col gap-2 max-w-lg z-[999]">
			<For each={notifications()}>
				{notification => {
					const Icon = notificationIcons[notification.type];
					return (
						<button
							role="alert"
							class="alert cursor-pointer alert-vertical sm:alert-horizontal"
							classList={{
								[notificationClasses[notification.type]]: true,
							}}
							onClick={() => notification.onClose?.()}>
							<Show when={Icon}>
								<Icon class="w-5 h-5" />
							</Show>
							<div>
								<Show when={notification.title}>
									<h3 class="font-bold">{notification.title}</h3>
								</Show>
								<div class="text-xs">{notification.message}</div>
							</div>
							<Show when={notification.customContent}>{notification.customContent}</Show>
						</button>
					);
				}}
			</For>
		</div>
	);
};
