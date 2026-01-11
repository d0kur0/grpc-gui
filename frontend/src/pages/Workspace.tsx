import { Viewport } from "../components/Viewport";
import "./Workspace.css";
import { WorkspaceServicesMenu } from "../components/WorkspaceServicesMenu";
import { Tabs } from "../components/Tabs";

export const Workspace = () => {
	return (
		<Viewport subtitle="workspace">
			<div class="workspace">
				<div class="workspace__menu">
					<WorkspaceServicesMenu />
				</div>

				<div class="workspace__body">
					<Tabs />
				</div>
			</div>
		</Viewport>
	);
};
