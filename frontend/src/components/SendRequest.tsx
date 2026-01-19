import { $servers } from "../stores/servers";
import { createMemo } from "solid-js";
import styles from "./SendRequest.module.css";

export type SendRequestProps = {
	serverId: number;
	methodName: string;
	serviceName: string;
};

export const SendRequest = (props: SendRequestProps) => {
	const {servers} = $servers

	const server = createMemo(() => {
		return servers().find(server => server.server?.id === props.serverId)!;
	});


	return (
		<div class={styles.root}>
			<div class={styles.description}>
				<span class={styles.descriptionTitle}>{server().server?.name}</span> / 
				<span class={styles.descriptionTitle}>{props.serviceName}</span> / 
				<span class={styles.descriptionTitle}>{props.methodName}</span>
			</div>
		</div>
	);
};
