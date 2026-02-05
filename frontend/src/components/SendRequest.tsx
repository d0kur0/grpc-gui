import { $servers } from "../stores/servers";
import { createMemo, createSignal, For, Show, onMount } from "solid-js";
import styles from "./SendRequest.module.css";
import { JsonEditor } from "./JsonEditor";
import { DoGRPCRequest } from "../../bindings/grpc-gui/app";
import { $notifications, NotificationType } from "../stores/notifications";
import { $history } from "../stores/history";
import { IoCopy, IoDownload } from "solid-icons/io";
import { History } from "../../bindings/grpc-gui/internal/models/models";
import stripJsonComments from "strip-json-comments";

export type SendRequestProps = {
	serverId: number;
	methodName: string;
	serviceName: string;
	historyData?: History;
};

type RequestTab = "body" | "metadata" | "context";

type KeyValuePair = {
	id: string;
	key: string;
	value: string;
};

export const SendRequest = (props: SendRequestProps) => {
	const { servers } = $servers;

	const server = createMemo(() => {
		return servers().find(server => server.server?.id === props.serverId)!;
	});

	const method = createMemo(() => {
		const srv = server();
		const service = srv.reflection?.services?.find(s => s.name === props.serviceName);
		return service?.methods?.find(m => m.name === props.methodName);
	});

	const [activeTab, setActiveTab] = createSignal<RequestTab>("body");
	const [requestBody, setRequestBody] = createSignal("{}");
	const [metadata, setMetadata] = createSignal<KeyValuePair[]>([
		{ id: crypto.randomUUID(), key: "", value: "" }
	]);
	const [contextValues, setContextValues] = createSignal<KeyValuePair[]>([
		{ id: crypto.randomUUID(), key: "", value: "" }
	]);
	const [response, setResponse] = createSignal<string>("");
	const [responseTime, setResponseTime] = createSignal<number>(0);
	const [isLoading, setIsLoading] = createSignal(false);

	onMount(() => {
		// Если есть данные из истории, загружаем их
		if (props.historyData) {
			const history = props.historyData;
			
			// Загружаем requestBody
			if (history.request) {
				try {
					const parsed = JSON.parse(history.request);
					setRequestBody(JSON.stringify(parsed, null, 2));
				} catch {
					setRequestBody(history.request);
				}
			}
			
			// Загружаем metadata
			if (history.requestHeaders) {
				try {
					const parsed = JSON.parse(history.requestHeaders);
					const metadataArray = Object.entries(parsed).map(([key, value]) => ({
						id: crypto.randomUUID(),
						key,
						value: String(value)
					}));
					if (metadataArray.length > 0) {
						setMetadata(metadataArray);
					}
				} catch (err) {
					console.error("Failed to parse request headers:", err);
				}
			}
			
			// Загружаем contextValues
			if (history.contextValues) {
				try {
					const parsed = JSON.parse(history.contextValues);
					const contextArray = Object.entries(parsed).map(([key, value]) => ({
						id: crypto.randomUUID(),
						key,
						value: String(value)
					}));
					if (contextArray.length > 0) {
						setContextValues(contextArray);
					}
				} catch (err) {
					console.error("Failed to parse context values:", err);
				}
			}
			
			// Загружаем response
			if (history.response) {
				try {
					const parsed = JSON.parse(history.response);
					const prettyJson = JSON.stringify(parsed, null, 2);
					const startTimeStr = new Date(history.createdAt).toLocaleString('ru-RU');
					const formattedResponse = `// Время начала выполнения запроса: ${startTimeStr}\n// Время выполнения запроса: ${history.executionTime}ms\n// Код ответа: ${history.statusCode}\n\n${prettyJson}`;
					setResponse(formattedResponse);
				} catch {
					const startTimeStr = new Date(history.createdAt).toLocaleString('ru-RU');
					const formattedResponse = `// Время начала выполнения запроса: ${startTimeStr}\n// Время выполнения запроса: ${history.executionTime}ms\n// Код ответа: ${history.statusCode}\n\n${history.response}`;
					setResponse(formattedResponse);
				}
			}
		} else {
			// Иначе загружаем пример из рефлексии
			const m = method();
			if (m?.requestExampleString) {
				setRequestBody(m.requestExampleString);
			} else if (m?.requestExample) {
				try {
					const exampleJson = JSON.stringify(m.requestExample, null, 2);
					setRequestBody(exampleJson);
				} catch (err) {
					console.error("Failed to parse request example:", err);
				}
			}
		}
	});

	const addMetadataRow = () => {
		setMetadata([...metadata(), { id: crypto.randomUUID(), key: "", value: "" }]);
	};

	const removeMetadataRow = (id: string) => {
		setMetadata(metadata().filter(item => item.id !== id));
	};

	const updateMetadataKey = (id: string, key: string) => {
		setMetadata(metadata().map(item => item.id === id ? { ...item, key } : item));
	};

	const updateMetadataValue = (id: string, value: string) => {
		setMetadata(metadata().map(item => item.id === id ? { ...item, value } : item));
	};

	const addContextRow = () => {
		setContextValues([...contextValues(), { id: crypto.randomUUID(), key: "", value: "" }]);
	};

	const removeContextRow = (id: string) => {
		setContextValues(contextValues().filter(item => item.id !== id));
	};

	const updateContextKey = (id: string, key: string) => {
		setContextValues(contextValues().map(item => item.id === id ? { ...item, key } : item));
	};

	const updateContextValue = (id: string, value: string) => {
		setContextValues(contextValues().map(item => item.id === id ? { ...item, value } : item));
	};

	const handleCopyResponse = () => {
		const textarea = document.createElement("textarea");
		textarea.value = response();
		textarea.style.position = "fixed";
		textarea.style.opacity = "0";
		document.body.appendChild(textarea);
		textarea.select();
		
		try {
			document.execCommand("copy");
			$notifications.addNotification({
				type: NotificationType.SUCCESS,
				title: "Скопировано",
				message: "Ответ скопирован в буфер обмена",
			});
		} catch (err) {
			$notifications.addNotification({
				type: NotificationType.ERROR,
				title: "Ошибка",
				message: "Не удалось скопировать",
			});
		} finally {
			document.body.removeChild(textarea);
		}
	};

	const handleDownloadResponse = () => {
		const blob = new Blob([response()], { type: "application/json" });
		const url = URL.createObjectURL(blob);
		const a = document.createElement("a");
		a.href = url;
		a.download = `response-${props.serviceName}-${props.methodName}-${Date.now()}.json`;
		document.body.appendChild(a);
		a.click();
		document.body.removeChild(a);
		URL.revokeObjectURL(url);
	};

	const handleSendRequest = async () => {
		try {
			setIsLoading(true);
			setResponse("");
			setResponseTime(0);

			const startTime = new Date();
			const startTimeStr = startTime.toLocaleString('ru-RU');

			const metadataObj: { [key: string]: string } = {};
			metadata().forEach(item => {
				if (item.key.trim()) {
					metadataObj[item.key] = item.value;
				}
			});

			const contextObj: { [key: string]: string } = {};
			contextValues().forEach(item => {
				if (item.key.trim()) {
					contextObj[item.key] = item.value;
				}
			});

			const cleanedRequestBody = stripJsonComments(requestBody());
			
			const [responseData, time] = await DoGRPCRequest(
				props.serverId,
				server().server?.address || "",
				props.serviceName,
				props.methodName,
				cleanedRequestBody,
				Object.keys(metadataObj).length > 0 ? metadataObj : null,
				Object.keys(contextObj).length > 0 ? contextObj : null
			);

			let formattedResponse = "";
			try {
				const parsed = JSON.parse(responseData);
				const prettyJson = JSON.stringify(parsed, null, 2);
				formattedResponse = `// Время начала выполнения запроса: ${startTimeStr}\n// Время выполнения запроса: ${time}ms\n// Код ответа: 0 (OK)\n\n${prettyJson}`;
			} catch {
				formattedResponse = `// Время начала выполнения запроса: ${startTimeStr}\n// Время выполнения запроса: ${time}ms\n// Код ответа: 0 (OK)\n\n${responseData}`;
			}

			setResponse(formattedResponse);
			setResponseTime(time);
			
			$history.refresh();
		} catch (err: any) {
			let errorMessage = err?.message || "Неизвестная ошибка";
			
			try {
				const parsed = JSON.parse(errorMessage);
				errorMessage = JSON.stringify(parsed, null, 2);
			} catch {
				
			}
			
			const errorResponse = `// Ошибка выполнения запроса\n// Код ответа: 1 (ERROR)\n\n${errorMessage}`;
			setResponse(errorResponse);
			
			$history.refresh();
					} finally {
			setIsLoading(false);
		}
	};

	return (
		<div class={styles.root}>
			<div class={styles.header}>
				<div class={styles.description}>
					<span class={styles.descriptionTitle}>{server().server?.name}</span>
					<span class={styles.separator}>/</span>
					<span class={styles.descriptionTitle}>{props.serviceName}</span>
					<span class={styles.separator}>/</span>
					<span class={styles.descriptionTitle}>{props.methodName}</span>
				</div>
				<div class={styles.address}>{server().server?.address}</div>
			</div>

			<div class={styles.content}>
				<div class={styles.requestSection}>
					<div class={styles.tabs}>
						<div class={styles.tabsList}>
							<button
								class={styles.tab}
								classList={{ [styles.tabActive]: activeTab() === "body" }}
								onClick={() => setActiveTab("body")}
							>
								Тело запроса
							</button>
							<button
								class={styles.tab}
								classList={{ [styles.tabActive]: activeTab() === "metadata" }}
								onClick={() => setActiveTab("metadata")}
							>
								Метаданные
							</button>
							<button
								class={styles.tab}
								classList={{ [styles.tabActive]: activeTab() === "context" }}
								onClick={() => setActiveTab("context")}
							>
								Контекст
							</button>
						</div>
						<button
							class="btn btn-sm btn-success"
							onClick={handleSendRequest}
							disabled={isLoading()}
						>
							{isLoading() ? "Отправка..." : "Отправить"}
						</button>
					</div>

					<div class={styles.tabContent}>
					<Show when={activeTab() === "body"}>
						<div class={styles.editorWrapper}>
							<JsonEditor 
								value={requestBody()} 
								onChange={setRequestBody}
								schema={method()?.request}
							/>
						</div>
					</Show>

					<Show when={activeTab() === "metadata"}>
						<div class={styles.keyValueList}>
							<div class={styles.keyValueDescription}>
								Метаданные передаются как gRPC заголовки (headers) в запросе
							</div>
							<For each={metadata()}>
								{(item) => (
									<div class={styles.keyValueRow}>
										<input
											type="text"
											class="input input-sm"
											placeholder="Key"
											value={item.key}
											onInput={(e) => updateMetadataKey(item.id, e.currentTarget.value)}
										/>
										<input
											type="text"
											class="input input-sm"
											placeholder="Value"
											value={item.value}
											onInput={(e) => updateMetadataValue(item.id, e.currentTarget.value)}
										/>
										<button
											class="btn btn-sm btn-ghost"
											onClick={() => removeMetadataRow(item.id)}
										>
											×
										</button>
									</div>
								)}
							</For>
							<div class={styles.keyValueActions}>
								<button class="btn btn-sm btn-neutral" onClick={addMetadataRow}>
									Добавить
								</button>
							</div>
						</div>
					</Show>

					<Show when={activeTab() === "context"}>
						<div class={styles.keyValueList}>
							<div class={styles.keyValueDescription}>
								Контекстные значения передаются в context.Context запроса
							</div>
							<For each={contextValues()}>
								{(item) => (
									<div class={styles.keyValueRow}>
										<input
											type="text"
											class="input input-sm"
											placeholder="Key"
											value={item.key}
											onInput={(e) => updateContextKey(item.id, e.currentTarget.value)}
										/>
										<input
											type="text"
											class="input input-sm"
											placeholder="Value"
											value={item.value}
											onInput={(e) => updateContextValue(item.id, e.currentTarget.value)}
										/>
										<button
											class="btn btn-sm btn-ghost"
											onClick={() => removeContextRow(item.id)}
										>
											×
										</button>
									</div>
								)}
							</For>
							<div class={styles.keyValueActions}>
								<button class="btn btn-sm btn-neutral" onClick={addContextRow}>
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
								disabled={!response()}
							>
								<IoCopy class="w-4 h-4" />
							</button>
							<button
								class="btn btn-xs btn-ghost"
								onClick={handleDownloadResponse}
								title="Скачать"
								disabled={!response()}
							>
								<IoDownload class="w-4 h-4" />
							</button>
						</div>
					</div>
					<div class={styles.editorWrapper}>
						<JsonEditor value={response()} readOnly={true} />
					</div>
				</div>
			</div>
		</div>
	);
};
