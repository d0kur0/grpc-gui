import "./App.css";
import "@wailsio/runtime";
import { Route, Router } from "@solidjs/router";
import { Workspace } from "./pages/Workspace"; 
import { AddService } from "./pages/AddService";
import { History } from "./pages/History";

export const App = () => {
	return (
		<Router>
			<Route path="/" component={Workspace} />
			<Route path="/add-service" component={AddService} />
			<Route path="/history" component={History} />
		</Router>
	);
};
