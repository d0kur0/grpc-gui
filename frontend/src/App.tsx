import "./App.css";
import { Viewport } from "./components/Viewport";
import "@wailsio/runtime";
import { Route, Router } from "@solidjs/router";
import { Workspace } from "./pages/Workspace";
import { AddService } from "./pages/AddService";

export const App = () => {
	return (
		<Router>
			<Route path="/" component={Workspace} />
			<Route path="/add-service" component={AddService} />
		</Router>
	);
};
