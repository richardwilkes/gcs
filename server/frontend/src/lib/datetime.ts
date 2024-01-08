const stdDateTimeFmt = new Intl.DateTimeFormat('default', {
	month: 'short',
	day: 'numeric',
	year: 'numeric',
	hour: 'numeric',
	minute: '2-digit',
	hour12: true
});

export function formatDateStamp(date: Date | string) {
	if (typeof date === 'string') {
		date = new Date(date);
	}
	return stdDateTimeFmt.format(date);
}
