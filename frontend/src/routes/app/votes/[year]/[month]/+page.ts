export function load({ data, params }) {
	const { isDefault, votes } = data;

	return {
		year: Number(params.year),
		month: params.month,
		isDefault,
		votes
	};
}
