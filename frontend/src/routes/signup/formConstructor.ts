import { type FormControl } from '$lib/types/forms';

export const address: readonly FormControl[] = [
	{
		name: 'city',
		label: 'Ville',
		type: 'text',
		placeholder: 'Entrer votre ville'
	},
	{
		name: 'postalcode',
		label: 'Code postal',
		type: 'text',
		placeholder: 'Entrer votre code postal'
	},
	{
		name: 'address1',
		label: 'Addresse 1',
		type: 'text',
		placeholder: 'Entrer le premier champ de votre addresse',
		isOneLine: true
	},
	{
		name: 'address2',
		label: 'Addresse 2',
		type: 'text',
		placeholder: 'Entrer le second champ de votre addresse',
		isOneLine: true
	}
] as const;

export const general: readonly FormControl[] = [
	{
		name: 'email',
		label: 'Email',
		type: 'email',
		placeholder: 'Entrer votre email'
	},
	{
		name: 'password',
		label: 'Mot de passe',
		type: 'password',
		placeholder: 'Entrer votre mot de passe'
	},
	{
		name: 'firstname',
		label: 'Prenom',
		type: 'text',
		placeholder: 'Entrer votre prenom'
	},
	{
		name: 'lastname',
		label: 'Nom',
		type: 'text',
		placeholder: 'Entrer votre nom'
	}
] as const;
