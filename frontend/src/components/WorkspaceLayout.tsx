import { JSX, createSignal, onMount } from "solid-js";
import "./WorkspaceLayout.css";

interface WorkspaceLayoutProps {
	menu: JSX.Element;
	body: JSX.Element;
}

export const WorkspaceLayout = (props: WorkspaceLayoutProps) => {
	const MIN_WIDTH = 300;
	const MAX_WIDTH = 550;
	const DEFAULT_WIDTH = 250;

	const [menuWidth, setMenuWidth] = createSignal(DEFAULT_WIDTH);
	const [isResizing, setIsResizing] = createSignal(false);

	onMount(() => {
		const savedWidth = localStorage.getItem("workspace-menu-width");
		if (savedWidth) {
			const width = parseInt(savedWidth, 10);
			if (width >= MIN_WIDTH && width <= MAX_WIDTH) {
				setMenuWidth(width);
			}
		}
	});

	const handleMouseDown = (e: MouseEvent) => {
		e.preventDefault();
		setIsResizing(true);

		const handleMouseMove = (e: MouseEvent) => {
			const newWidth = Math.min(Math.max(e.clientX, MIN_WIDTH), MAX_WIDTH);
			setMenuWidth(newWidth);
		};

		const handleMouseUp = () => {
			setIsResizing(false);
			localStorage.setItem("workspace-menu-width", menuWidth().toString());
			document.removeEventListener("mousemove", handleMouseMove);
			document.removeEventListener("mouseup", handleMouseUp);
		};

		document.addEventListener("mousemove", handleMouseMove);
		document.addEventListener("mouseup", handleMouseUp);
	};

	return (
		<div class="workspace-layout" classList={{ "workspace-layout-resizing": isResizing() }}>
			<div class="workspace-layout__menu scrollbar" style={{ width: `${menuWidth()}px` }}>
				{props.menu}
			</div>

			<div class="workspace-layout__resizer" onMouseDown={handleMouseDown} />

			<div class="workspace-layout__body">{props.body}</div>
		</div>
	);
};
