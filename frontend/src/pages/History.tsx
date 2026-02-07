import { Viewport } from "../components/Viewport";
import { HistoryList } from "../components/HistoryList";
import { Tabs } from "../components/Tabs";
import { WorkspaceLayout } from "../components/WorkspaceLayout";
import "./Workspace.css";

export const History = () => {
	return (
		<Viewport subtitle="history">
			<WorkspaceLayout menu={<HistoryList />} body={<Tabs />} />
		</Viewport>
	);
};
