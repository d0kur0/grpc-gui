import { createSignal } from "solid-js";
import { Viewport } from "../components/Viewport";
import { ValidationStatus } from "../../bindings/grpc-gui/models";
import { CreateServer, ValidateServerAddress } from "../../bindings/grpc-gui/app";
import { $notifications, NotificationType } from "../stores/notifications";
import { $servers } from "../stores/servers";

export const AddService = () => {
	const [isLoading, setIsLoading] = createSignal(false);
	const [loadingMessage, setLoadingMessage] = createSignal("");

	const handleSubmit = async (e: SubmitEvent) => {
		e.preventDefault();
		const formData = new FormData(e.target as HTMLFormElement);
		const name = formData.get("name") as string;
		const address = formData.get("address") as string;
		const useTLS = formData.get("optUseTLS") === "on";
		const insecure = formData.get("optInsecure") === "on";

		try {
			setIsLoading(true);
			setLoadingMessage("Проверка адреса сервиса...");
			const validationResult = await ValidateServerAddress(address, useTLS, insecure);
			if (validationResult.status !== ValidationStatus.ValidationStatusSuccess) {
				const errorMessage = validationResult.message || "Ошибка валидации сервера";
				throw new Error(errorMessage);
			}
			setLoadingMessage("Добавление сервиса...");
			await CreateServer(name, address, useTLS, insecure);
			$notifications.addNotification({
				message: "Сервис успешно добавлен",
				title: "Успех",
				type: NotificationType.SUCCESS,
			});
			await $servers.refreshServers();
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

	return (
		<Viewport subtitle="новый сервис">
			<div class="px-6 py-4">
				<div class="text-xl font-bold">Добавление сервиса</div>
				<div class="text-sm text-base-content/50 my-2 max-w-xl">
					Вводишь тут адрес GRPC сервиса
					<br />
					На базе рефлексии достанем все сервисы, методы и это вот все, возможно потом, для сервисов, что не
					предоставляют рефлексию, можно будет залить protobuf файлы, но пока похуй
				</div>

				<form onSubmit={handleSubmit} class="w-full mt-6 space-y-1">
					<fieldset class="fieldset">
						<legend class="fieldset-legend">Название сервиса</legend>
						<input
							required
							type="text"
							class="input max-w-md w-full"
							placeholder="bullshit API v1"
							name="name"
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
							name="address"
						/>
						<p class="label">Например: localhost:50051, базовый URL для запросов. стейджингов</p>
					</fieldset>

					<div class="flex items-center gap-4">
						<fieldset class="fieldset bg-base-100 border-base-300 rounded-box w-64 border p-4">
							<legend class="fieldset-legend">TLS</legend>
							<label class="label">
								<input type="checkbox" checked={true} class="toggle" name="optUseTLS" />
								Использовать TLS?
							</label>
						</fieldset>

						<fieldset class="fieldset bg-base-100 border-base-300 rounded-box w-64 border p-4">
							<legend class="fieldset-legend">Без проверки сертификата</legend>
							<label class="label">
								<input type="checkbox" checked={true} class="toggle" name="optInsecure" />
								insecure skip verify?
							</label>
						</fieldset>
					</div>

					<div class="text-sm text-base-content/60 max-w-3xl my-6 p-5  border-base-content/20 rounded-box">
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

					<button class="btn btn-primary" disabled={isLoading()}>
						{isLoading() ? loadingMessage() : "Схоронить"}
					</button>
				</form>
			</div>
		</Viewport>
	);
};
