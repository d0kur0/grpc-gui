import { createEffect, createSignal, JSX, Show } from "solid-js";
import "./Viewport.css";
import { Window } from "@wailsio/runtime";
import { RiDevelopmentTerminalBoxFill } from "solid-icons/ri";
import { FaSolidArrowRightLong } from "solid-icons/fa";

type ViewportProps = {
	subtitle?: string;
	children: JSX.Element;
};

export const Viewport = (props: ViewportProps) => {
	const [isMaximised, setIsMaximised] = createSignal(false);

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
		<div class="viewport">
			<div class="viewport__controls">
				<div class="viewport__title">
					grpc-gui
					<Show when={props.subtitle}>
						<FaSolidArrowRightLong />
						<span class="viewport__subtitle">{props.subtitle}</span>
					</Show>
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
		</div>
	);
};
