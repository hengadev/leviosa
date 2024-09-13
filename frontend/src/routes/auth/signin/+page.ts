import type { PageLoad } from './$types';
import { type FormControl, type OAuthButtons } from '$lib/types/forms';

import Apple from '../../../assets/apple.svelte';
import Google from '../../../assets/google.svelte';
// import Instagram from '../../../assets/instagram.svelte';

// TODO: remove when app done, just for testing, should use FormControl
type TestingFormControl = FormControl & {
	value: string;
};

const formControls: TestingFormControl[] = [
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

// NOTE: for the moment I cannot use the appli button and instagram does not provide oauth
const oauthButtons: OAuthButtons[] = [
	{
		name: 'Apple',
		providerName: 'apple',
		IconComponent: Apple
	},
	{
		name: 'Google',
		providerName: 'google',
		IconComponent: Google
	}
	// {
	// 	name: 'Instagram',
	// 	providerName: 'instagram',
	// 	IconComponent: Instagram
	// }
];

export const load: PageLoad = ({ data }) => {
	const url = data.API_URL;
	return { formControls, oauthButtons, url };
};
