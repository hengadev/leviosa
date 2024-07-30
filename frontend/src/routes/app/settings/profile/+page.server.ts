import type { Actions, PageServerLoad } from "./$types"
import { redirect } from "@sveltejs/kit"

const missingValues = {
    email: "Aucune adresse email précisé",
    telephone: "Aucun numero de telephone précisé",
    lastname: "Aucun nom précisé",
    firstname: "Aucun prenom précisé",
    birthdate: "Aucune date de naissance précisé",
    address: "Aucune adresse précisé",
    city: "Aucune ville précisé",
    postalcard: "Aucun code postal précisé",
} as const

// TODO:
export const load: PageServerLoad = ({ locals }) => {
    if (!locals.user) {
        console.log("redirect")
        throw redirect(301, "/")
    }

    const user = locals.user;
    return { missingValues, user }
}

export const actions: Actions = {
    default: async ({ request, fetch }) => {
        const formData = request.formData()
        const fieldName = formData.get("fieldName")
        // TODO: Use that to make the put request to update the field if the user change it.
    }
}
