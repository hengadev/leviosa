import type { PageLoad } from './$types';
import { type FormControl } from '$lib/types/forms';

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

export const load: PageLoad = () => {
    return { formControls };
};
