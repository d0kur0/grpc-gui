import { createSignal, Show } from "solid-js";
import { ServerWithReflection } from "../../bindings/grpc-gui";
import { CreateServer, UpdateServer, ValidateServerAddress } from "../../bindings/grpc-gui/app";
import { $notifications, NotificationType } from "../stores/notifications";
import { ValidationStatus } from "../../bindings/grpc-gui/models";

export type ServerModalProps = {
	server?: ServerWithReflection;
	onClose: () => void;
	onSuccess: (serverId?: number, isEdit?: boolean) => void;
};

export const ServerModal = (props: ServerModalProps) => {
	const isEdit = () => !!props.server;

	const [form, setForm] = createSignal({
		name: props.server?.server?.name || "",
		address: props.server?.server?.address || "",
		useTLS: props.server?.server?.optUseTLS ?? true,
		insecure: props.server?.server?.optInsecure ?? true,
	});

	const [isLoading, setIsLoading] = createSignal(false);
	const [loadingMessage, setLoadingMessage] = createSignal("");

	const handleSubmit = async (e: SubmitEvent) => {
		e.preventDefault();
		const f = form();

		try {
			setIsLoading(true);
			setLoadingMessage("Проверка адреса сервиса...");
			const validationResult = await ValidateServerAddress(f.address, f.useTLS, f.insecure);
			if (validationResult.status !== ValidationStatus.ValidationStatusSuccess) {
				const errorMessage = validationResult.message || "Ошибка валидации сервера";
				throw new Error(errorMessage);
			}

			let serverId: number | undefined;

			if (isEdit()) {
				serverId = props.server!.server!.id;
				setLoadingMessage("Обновление сервиса...");
				await UpdateServer(serverId, f.name, f.address, f.useTLS, f.insecure);
				$notifications.addNotification({
					message: "Сервис успешно обновлен",
					title: "Успех",
					type: NotificationType.SUCCESS,
				});
			} else {
				setLoadingMessage("Добавление сервиса...");
				serverId = await CreateServer(f.name, f.address, f.useTLS, f.insecure);
				$notifications.addNotification({
					message: "Сервис успешно добавлен",
					title: "Успех",
					type: NotificationType.SUCCESS,
				});
			}

			props.onSuccess(serverId, isEdit());
			props.onClose();
		} catch (error) {
			$notifications.addNotification({
				message: error instanceof Error ? error.message : "Неизвестная ошибка",
				title: "Ошибка",
				type: NotificationType.ERROR,
			});
		} finally {
			setIsLoading(false);
			setLoadingMessage("");
		}
	};

	const handleClose = () => {
		if (isLoading()) return;
		props.onClose();
	};

	return (
		<dialog class="modal modal-open" onClick={handleClose}>
			<div class="modal-box max-w-4xl" onClick={e => e.stopPropagation()}>
				<h3 class="text-xl font-bold mb-2">{isEdit() ? "Редактирование сервиса" : "Добавление сервиса"}</h3>
				<div class="text-sm text-base-content/50 mb-6">
					{isEdit() ? "Обнови данные сервиса и сохрани" : "Вводишь тут адрес GRPC сервиса"}
					<Show when={!isEdit()}>
						<br />
						На базе рефлексии достанем все сервисы, методы и это вот все, возможно потом, для сервисов, что не
						предоставляют рефлексию, можно будет залить protobuf файлы, но пока похуй
					</Show>
				</div>

				<form onSubmit={handleSubmit} class="space-y-1">
					<fieldset class="fieldset">
						<legend class="fieldset-legend">Название сервиса</legend>
						<input
							required
							type="text"
							class="input max-w-md w-full"
							placeholder="bullshit API v1"
							value={form().name}
							onInput={e => setForm({ ...form(), name: e.currentTarget.value })}
						/>
						<p class="label">Просто человекочитаемое название</p>
					</fieldset>

					<fieldset class="fieldset">
						<legend class="fieldset-legend">Адрес GRPC сервиса</legend>
						<input
							required
							type="text"
							class="input max-w-md w-full"
							placeholder="localhost:50051"
							value={form().address}
							onInput={e => setForm({ ...form(), address: e.currentTarget.value })}
						/>
						<p class="label">Например: localhost:50051, базовый URL для запросов. стейджингов</p>
					</fieldset>

					<div class="flex items-center gap-4">
						<fieldset class="fieldset bg-base-100 border-base-300 rounded-box w-64 border p-4">
							<legend class="fieldset-legend">TLS</legend>
							<label class="label" onClick={e => e.stopPropagation()}>
								<input
									type="checkbox"
									class="toggle"
									checked={form().useTLS}
									onChange={e => setForm({ ...form(), useTLS: e.currentTarget.checked })}
								/>
								Использовать TLS?
							</label>
						</fieldset>

						<fieldset class="fieldset bg-base-100 border-base-300 rounded-box w-64 border p-4">
							<legend class="fieldset-legend">Без проверки сертификата</legend>
							<label class="label" onClick={e => e.stopPropagation()}>
								<input
									type="checkbox"
									class="toggle"
									checked={form().insecure}
									onChange={e => setForm({ ...form(), insecure: e.currentTarget.checked })}
								/>
								insecure skip verify?
							</label>
						</fieldset>
					</div>

					<div class="text-sm text-base-content/60 max-w-3xl my-6 p-5 border-1 border-dashed border-base-content/20 rounded-box">
						Если включаешь TLS, то юзается пустой{" "}
						<span class="text-base-content/90 text-xs font-mono">{`credentials.NewTLS(&tls.Config{})`}</span>
						<br />
						Если дополнительно включишь без проверки сертификата, то юзается{" "}
						<span class="text-base-content/90 text-xs font-mono">{`credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})`}</span>
						, иначе используется{" "}
						<span class="text-base-content/90 text-xs font-mono">{`insecure.NewCredentials()`}</span>
						<br />
						Вообще, тут наверное нужно будет дать возможность залить кастомый сертификат и ключ, но пока похуй
					</div>

					<div class="modal-action">
						<button type="button" class="btn" onClick={handleClose} disabled={isLoading()}>
							Отмена
						</button>
						<button type="submit" class="btn btn-primary" disabled={isLoading()}>
							{isLoading() ? loadingMessage() : "Схоронить"}
						</button>
					</div>
				</form>
			</div>
		</dialog>
	);
};
