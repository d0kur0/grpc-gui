import { createSignal } from "solid-js";
import { Viewport } from "../components/Viewport";
import { Server } from "../../bindings/grpc-gui/internal/models";

export const AddService = () => {
	const [formFields, setFormFields] = createSignal<Server>({
		id: 0,
		createdAt: new Date(),
		updatedAt: new Date(),
		deletedAt: null,
		name: "",
		address: "",
		optUseTLS: false,
		optInsecure: false,
	});

	const handleSubmit = (e: SubmitEvent) => {
		e.preventDefault();
		const formData = new FormData(e.target as HTMLFormElement);
		const formFields = Object.fromEntries(formData.entries());

		console.log(formFields);
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
						<input type="text" class="input max-w-md w-full" placeholder="bullshit API v1" />
						<p class="label">Просто человекочитаемое название</p>
					</fieldset>

					<fieldset class="fieldset">
						<legend class="fieldset-legend">Адрес GRPC сервиса</legend>
						<input type="text" class="input max-w-md w-full" placeholder="localhost:50051" />
						<p class="label">Например: localhost:50051, базовый URL для запросов. стейджингов</p>
					</fieldset>

					<div class="flex items-center gap-4">
						<fieldset class="fieldset bg-base-100 border-base-300 rounded-box w-64 border p-4">
							<legend class="fieldset-legend">TLS</legend>
							<label class="label">
								<input type="checkbox" checked={true} class="toggle" />
								Использовать TLS?
							</label>
						</fieldset>

						<fieldset class="fieldset bg-base-100 border-base-300 rounded-box w-64 border p-4">
							<legend class="fieldset-legend">Без проверки сертификата</legend>
							<label class="label">
								<input type="checkbox" checked={true} class="toggle" />
								insecure skip verify?
							</label>
						</fieldset>
					</div>

					<div class="text-sm text-base-content/60 max-w-3xl my-6 p-5 border-1 border-base-content/20 rounded-box">
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

					<button class="btn btn-primary">Схоронить</button>
				</form>
			</div>
		</Viewport>
	);
};
