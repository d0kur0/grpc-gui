import "./App.css";
import "@wailsio/runtime";
import { Route, Router } from "@solidjs/router";
import { Workspace } from "./pages/Workspace"; 
import { History } from "./pages/History";
import { AddService } from "./pages/AddService";

export const App = () => {
	return (
		<Router>
			<Route path="/" component={Workspace} />
			<Route path="/history" component={History} />
			<Route path="/add-service" component={AddService} />
		</Router>
	);
};
