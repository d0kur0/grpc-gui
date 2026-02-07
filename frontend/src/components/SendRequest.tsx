import { $servers } from "../stores/servers";
import { createMemo, createSignal, For, Show, onMount } from "solid-js";
import styles from "./SendRequest.module.css";
import { JsonEditor } from "./JsonEditor";
import { $tabs, SendRequestData } from "../stores/tabs";
import { IoCopy, IoDownload } from "solid-icons/io";

export type SendRequestProps = {
	tabId: string;
};

export const SendRequest = (props: SendRequestProps) => {
	const { servers } = $servers;
	const {
		tabs,
		updateTabData,
		sendRequest,
		addMetadataRow,
		removeMetadataRow,
		updateMetadataKey,
		updateMetadataValue,
		addContextRow,
		removeContextRow,
		updateContextKey,
		updateContextValue,
	} = $tabs;

	const tab = createMemo(() => tabs.find(t => t.id === props.tabId));
	const data = createMemo(() => tab()?.data as SendRequestData | undefined);

	const server = createMemo(() => {
		const d = data();
		if (!d) return undefined;
		return servers().find(s => s.server?.id === d.serverId);
	});

	const method = createMemo(() => {
		const d = data();
		const srv = server();
		if (!d || !srv) return undefined;
		const service = srv.reflection?.services?.find(s => s.name === d.serviceName);
		return service?.methods?.find(m => m.name === d.methodName);
	});

	onMount(() => {
		const d = data();
		if (!d || d.requestBody !== "{}" || d.historyData) return;

		const m = method();
		if (m?.requestExampleString) {
			updateTabData(props.tabId, { requestBody: m.requestExampleString });
		} else if (m?.requestExample) {
			try {
				const exampleJson = JSON.stringify(m.requestExample, null, 2);
				updateTabData(props.tabId, { requestBody: exampleJson });
			} catch (err) {
				console.error("Failed to parse request example:", err);
			}
		}
	});

	const [isLoading, setIsLoading] = createSignal(false);

	const handleCopyResponse = () => {
		const response = data()?.response;
		if (!response) return;

		const textarea = document.createElement("textarea");
		textarea.value = response;
		textarea.style.position = "fixed";
		textarea.style.opacity = "0";
		document.body.appendChild(textarea);
		textarea.select();

		try {
			document.execCommand("copy");
		} catch (err) {
		} finally {
			document.body.removeChild(textarea);
		}
	};

	const handleDownloadResponse = () => {
		const d = data();
		if (!d?.response) return;

		const blob = new Blob([d.response], { type: "application/json" });
		const url = URL.createObjectURL(blob);
		const a = document.createElement("a");
		a.href = url;
		a.download = `response-${d.serviceName}-${d.methodName}-${Date.now()}.json`;
		document.body.appendChild(a);
		a.click();
		document.body.removeChild(a);
		URL.revokeObjectURL(url);
	};

	const handleSendRequest = async () => {
		const srv = server();
		if (!srv?.server?.address) return;

		setIsLoading(true);
		try {
			await sendRequest(props.tabId, srv.server.address);
		} finally {
			setIsLoading(false);
		}
	};

	return (
		<Show when={data()}>
			{d => (
				<div class={styles.root}>
					<div class={styles.header}>
						<div class={styles.description}>
							<span class={styles.descriptionTitle}>{server()?.server?.name}</span>
							<span class={styles.separator}>/</span>
							<span class={styles.descriptionTitle}>{d().serviceName}</span>
							<span class={styles.separator}>/</span>
							<span class={styles.descriptionTitle}>{d().methodName}</span>
						</div>
						<div class={styles.address}>{server()?.server?.address}</div>
					</div>

					<div class={styles.content}>
						<div class={styles.requestSection}>
							<div class={styles.tabs}>
								<div class={styles.tabsList}>
									<button
										class={styles.tab}
										classList={{ [styles.tabActive]: d().activeTab === "body" }}
										onClick={() => updateTabData(props.tabId, { activeTab: "body" })}>
										Тело запроса
									</button>
									<button
										class={styles.tab}
										classList={{ [styles.tabActive]: d().activeTab === "metadata" }}
										onClick={() => updateTabData(props.tabId, { activeTab: "metadata" })}>
										Метаданные
									</button>
									<button
										class={styles.tab}
										classList={{ [styles.tabActive]: d().activeTab === "context" }}
										onClick={() => updateTabData(props.tabId, { activeTab: "context" })}>
										Контекст
									</button>
								</div>
								<button class="btn btn-sm btn-success" onClick={handleSendRequest} disabled={isLoading()}>
									{isLoading() ? "Отправка..." : "Отправить"}
								</button>
							</div>

							<div class={styles.tabContent}>
								<Show when={d().activeTab === "body"}>
									<div class={styles.editorWrapper}>
										<JsonEditor
											initialValue={d().requestBody}
											onChange={value => updateTabData(props.tabId, { requestBody: value })}
											schema={method()?.request}
										/>
									</div>
								</Show>

								<Show when={d().activeTab === "metadata"}>
									<div class={styles.keyValueList}>
										<div class={styles.keyValueDescription}>
											Метаданные передаются как gRPC заголовки (headers) в запросе
										</div>
										<For each={d().metadata}>
											{item => (
												<div class={styles.keyValueRow}>
													<input
														type="text"
														class="input input-sm"
														placeholder="Key"
														value={item.key}
														onInput={e => updateMetadataKey(props.tabId, item.id, e.currentTarget.value)}
													/>
													<input
														type="text"
														class="input input-sm"
														placeholder="Value"
														value={item.value}
														onInput={e => updateMetadataValue(props.tabId, item.id, e.currentTarget.value)}
													/>
													<button class="btn btn-sm btn-ghost" onClick={() => removeMetadataRow(props.tabId, item.id)}>
														×
													</button>
												</div>
											)}
										</For>
										<div class={styles.keyValueActions}>
											<button class="btn btn-sm btn-neutral" onClick={() => addMetadataRow(props.tabId)}>
												Добавить
											</button>
										</div>
									</div>
								</Show>

								<Show when={d().activeTab === "context"}>
									<div class={styles.keyValueList}>
										<div class={styles.keyValueDescription}>
											Контекстные значения передаются в context.Context запроса
										</div>
										<For each={d().contextValues}>
											{item => (
												<div class={styles.keyValueRow}>
													<input
														type="text"
														class="input input-sm"
														placeholder="Key"
														value={item.key}
														onInput={e => updateContextKey(props.tabId, item.id, e.currentTarget.value)}
													/>
													<input
														type="text"
														class="input input-sm"
														placeholder="Value"
														value={item.value}
														onInput={e => updateContextValue(props.tabId, item.id, e.currentTarget.value)}
													/>
													<button class="btn btn-sm btn-ghost" onClick={() => removeContextRow(props.tabId, item.id)}>
														×
													</button>
												</div>
											)}
										</For>
										<div class={styles.keyValueActions}>
											<button class="btn btn-sm btn-neutral" onClick={() => addContextRow(props.tabId)}>
												Добавить
											</button>
										</div>
									</div>
								</Show>
							</div>
						</div>

						<div class={styles.responseSection}>
							<div class={styles.responseHeader}>
								<span class={styles.responseTitle}>Ответ</span>
								<div class={styles.responseActions}>
									<button
										class="btn btn-xs btn-ghost"
										onClick={handleCopyResponse}
										title="Скопировать"
										disabled={!d().response}>
										<IoCopy class="w-4 h-4" />
									</button>
									<button
										class="btn btn-xs btn-ghost"
										onClick={handleDownloadResponse}
										title="Скачать"
										disabled={!d().response}>
										<IoDownload class="w-4 h-4" />
									</button>
								</div>
							</div>
							<div class={styles.editorWrapper}>
								<JsonEditor initialValue={d().response} readOnly={true} />
							</div>
						</div>
					</div>
				</div>
			)}
		</Show>
	);
};
