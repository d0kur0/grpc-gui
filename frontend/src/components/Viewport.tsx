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
					<button class="cursor-pointer" onClick={() => navigate("/")}>
						grpc-gui
					</button>
					<Show when={props.subtitle}>
						<FaSolidArrowRightLong />
						<span class="viewport__subtitle">{props.subtitle}</span>
					</Show>
				</div>

				<div class="viewport__app-controls">
					<button 
						title="Воркспейс" 
						class="btn btn-xs" 
						classList={{ "btn-neutral": location.pathname === "/" }}
						onClick={() => navigate("/")}
					>
						<IoHome class="w-4 h-4" />
					</button>
					<button 
						title="История" 
						class="btn btn-xs" 
						classList={{ "btn-neutral": location.pathname === "/history" }}
						onClick={() => navigate("/history")}
					>
						<IoTime class="w-4 h-4" />
					</button>
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
		<div class="absolute bottom-5 right-5 flex flex-col gap-2 max-w-lg">
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
