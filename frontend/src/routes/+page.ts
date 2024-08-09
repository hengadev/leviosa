import type { PageLoad } from './$types';

type InputType = 'email' | 'password' | 'text' | 'hidden';

type FormControl = {
	name: string;
	label: string;
	type: InputType;
	placeholder?: string;
	value: string; // just for now for the testing part with the backend
};
const formControls: FormControl[] = [
	{
		name: 'email',
		label: 'Email',
		type: 'email',
		placeholder: 'Entrer votre email',
		value: 'admin-livio@outlook.fr'
	},
	{
		name: 'password',
		label: 'Mot de passe',
		type: 'password',
		placeholder: 'Entrer votre mot de passe',
		value: 'secret1234'
	}
];

export const load: PageLoad = () => {
	// TODO: get to see if the cookie is set and if we do not need to reset it.
	// redirect if the cookie already set.
	// const a = 5;
	// if (a < 6) {
	//     throw redirect(300, "/app")
	// }
	return {
		formControls
	};
};
