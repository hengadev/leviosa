import { redirect } from '@sveltejs/kit';
import type { PageLoad } from './$types';

// TODO: Put all these types in a file where I can reuse them if necessary

// NOTE: A very strict country code in here to make
type CountryCode = 'fr' | 'com';
type Email = `${string}@${string}.${CountryCode}`;

type User = {
    email: Email;
    password: string;
};

type UserInfo = User & {
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
    placeholder?: FormControlPlaceholder;
};

const addressFormControls: readonly FormControl[] = [
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

const generalFormControls: readonly FormControl[] = [
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

// TODO: can i use the cookie in here.
// export const load: PageLoad = ({ cookies }) => {
export const load: PageLoad = () => {
    // TODO: do something to get to read if there is any valid cookie set and redirect if so
    // if (cookies.get("somecookie")) {
    //     redirect(300, "/app")
    // }
    return {
        general: generalFormControls,
        address: addressFormControls
    };
};
