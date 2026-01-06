import { createRoot, createSignal } from "solid-js";
import type { JSX } from "solid-js";

const DEFAULT_DURATION = 5000;
const MAX_NOTIFICATIONS_ON_SCREEN = 3;

export enum NotificationType {
	SUCCESS,
	ERROR,
	WARNING,
	INFO,
}

type Notification = {
	id: string;
	message: string;
	title?: string;
	type: NotificationType;
	duration?: number;
	onClose?: () => void;
	customContent?: JSX.Element;
};

const createNotificationsStore = () => {
	const [notifications, setNotifications] = createSignal<Notification[]>([]);

	const addNotification = (notification: Omit<Notification, "id">) => {
		const id = crypto.randomUUID();

		if (notifications().length >= MAX_NOTIFICATIONS_ON_SCREEN) {
			setNotifications(prev => prev.slice(1));
		}

		if (!notification.onClose) {
			notification.onClose = () => {
				setNotifications(prev => prev.filter(n => n.id !== id));
			};
		}

		if (notification.duration !== 0) {
			setTimeout(() => {
				setNotifications(prev => prev.filter(n => n.id !== id));
			}, notification.duration ?? DEFAULT_DURATION);
		}

		setNotifications(prev => [...prev, { ...notification, id }]);
	};

	return {
		notifications,
		addNotification,
	};
};

export const $notifications = createRoot(createNotificationsStore);
