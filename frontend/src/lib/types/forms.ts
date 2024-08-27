type CountryCode = 'fr' | 'com';
type Email = `${string}@${string}.${CountryCode}`;

type Credentials = {
	email: Email;
	password: string;
};

type UserInfo = Credentials & {
	city: string;
	postalcode: string;
	address1: string;
	address2: string;
	firstname: string;
	lastname: string;
};

type FormControlName = keyof UserInfo;
type FormControlInput = 'email' | 'password' | 'text' | 'hidden';
type FormControlPlaceholder = `Entrer ${string}`;

export type FormControl = {
	name: FormControlName;
	label: string;
	isOneLine?: boolean;
	type: FormControlInput;
	placeholder?: FormControlPlaceholder | string;
};
