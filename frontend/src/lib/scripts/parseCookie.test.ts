describe('parseCookie', () => {
	// just an example for the test in here and for me to use all my imports
	const cookie: string = '';
	it.todo('should do nothing for now', () => {
		const cookieParsed = parseCookie(cookie);
		// TODO: find the right format for the cookoie that I am going to exploit
		expect(cookieParsed).toHaveTextContent(/sometextcontent/iu);
	});
	expect(2 + 2).toBe(4);
});
