export type EmptyFallbackProps = {
	message: string;
};

export const EmptyFallback = (props: EmptyFallbackProps) => {
	return (
		<div class="flex items-center justify-center text-base-content/50	text-sm border-3 border-base-content/10 rounded-md p-4 border-dashed">
			{props.message}
		</div>
	);
};
