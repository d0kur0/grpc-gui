import { Viewport } from "../components/Viewport";
import { HistoryList } from "../components/HistoryList";
import { Tabs } from "../components/Tabs";
import "./Workspace.css";

export const History = () => {
	return (
		<Viewport subtitle="history">
			<div class="workspace">
				<div class="workspace__menu scrollbar">
					<HistoryList />
				</div>

				<div class="workspace__body">
					<Tabs />
				</div>
			</div>
		</Viewport>
	);
};
