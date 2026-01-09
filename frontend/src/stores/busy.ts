import { createRoot, createSignal } from "solid-js";
import { sleep } from "../utils/sleep";

type BusyParams = {
	startTimestamp: number;
	minimumDuration: number;
};

type BusyMap = Record<string, BusyParams | undefined>;

const createBusyStore = () => {
	const [busy, setBusy] = createSignal<BusyMap>({});

	const lockByKey = (key: string, minimumDuration: number = 0) => {
		setBusy(prev => ({ ...prev, [key]: { startTimestamp: Date.now(), minimumDuration } }));
	};

	const unlockByKey = async (key: string) => {
		const prev = busy()[key];

		if (prev) {
			const currentTimestamp = Date.now();
			const diff = currentTimestamp - prev.startTimestamp;

			if (diff < prev.minimumDuration) {
				await sleep(prev.minimumDuration - diff);
			}
		}

		setBusy(prev => ({ ...prev, [key]: undefined }));
	};

	return {
		busy,
		lockByKey,
		unlockByKey,
	};
};

export const $busy = createRoot(createBusyStore);
