type SendRequestProps = {
	serverId: number;
	methodName: string;
	serviceName: string;
};

export const SendRequest = (props: SendRequestProps) => {
	return (
		<div class="send-request">
			<div class="send-request__header">
				<div class="send-request__header-title">Send Request</div>
			</div>
		</div>
	);
};
