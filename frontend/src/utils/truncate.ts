export const truncateFromEnd = (text: string, maxLength: number): string => {
	if (text.length <= maxLength) return text;
	return "..." + text.slice(-maxLength);
};
