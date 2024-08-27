// TODO: use that video from kevin Powell to fix the user feedback: https://www.youtube.com/watch?v=awNYtIAu6pI
describe('validateEmail', () => {
	it.each([['someemail@gmail.com', '1234']])(
		'should invalid email shorter than X characters',
		(email: string, password: string) => {
			expect(async () => validate(email, password)).rejects.toThrowError();
		}
	);
	it('should validate email and password', () => {});
});

describe('validatePassword', () => {
	it.todo('should invalid password shorter than X characters');
});
