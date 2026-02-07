import { Viewport } from "../components/Viewport";
import "./Workspace.css";
import { WorkspaceServicesMenu } from "../components/WorkspaceServicesMenu";
import { Tabs } from "../components/Tabs";
import { WorkspaceLayout } from "../components/WorkspaceLayout";

export const Workspace = () => {
	return (
		<Viewport subtitle="workspace">
			<WorkspaceLayout menu={<WorkspaceServicesMenu />} body={<Tabs />} />
		</Viewport>
	);
};
