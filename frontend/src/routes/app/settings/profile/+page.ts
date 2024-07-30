//
import type { PageLoad } from "./$types"
import type { ComponentType } from "svelte"
import type { Icon } from "lucide-svelte"
// icons
import Mail from 'lucide-svelte/icons/mail';
import Phone from 'lucide-svelte/icons/phone';
import Mailbox from 'lucide-svelte/icons/mailbox';
import User from 'lucide-svelte/icons/user';
import Cake from 'lucide-svelte/icons/cake';


// TODO: get the user fields from the data user thing.
const userfields = ["email", "telephone", "lastname", "firstname", "birthdate", "address", "city", "postalcard"];
type UserFields = typeof userfields[number];
type FieldValueConstructor = {
    missingText: string,
    icon: ComponentType<Icon>
}
type Static = Record<UserFields, FieldValueConstructor>
type FieldValue = FieldValueConstructor & { value: string }
type Field = Record<UserFields, FieldValue>

const fieldsStatic: Static = {
    email: {
        missingText: "Email manquant",
        icon: Mail
    },
    telephone: {
        missingText: "Numero de telephone manquant",
        icon: Phone
    },
    lastname: {
        missingText: "Nom manquant",
        icon: User
    },
    firstname: {
        missingText: "Prenom manquant",
        icon: User
    },
    birthdate: {
        missingText: "Date de naissance manquante",
        icon: Cake
    },
    address: {
        missingText: "Adresse manquante",
        icon: Mailbox
    },
    city: {
        missingText: "Ville manquante",
        icon: Mailbox
    },
    postalcard: {
        missingText: "Code postal manquant",
        icon: Mailbox
    },
} as const

export const load: PageLoad = ({ data }) => {
    const { missingValues, user } = data
    const name: string = user.firstname + " " + user.lastname
    const fields: Field = {}
    for (const field of userfields) {
        const value = user[field] !== "" ? user[field] : missingValues[field]
        fields[field] = { ...fieldsStatic[field], value }
    }

    return { name, fields }
}
