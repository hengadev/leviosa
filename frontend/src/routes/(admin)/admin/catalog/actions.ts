import type { RequestEvent } from "./$types"
import { superValidate, setError } from 'sveltekit-superforms';
import { fail } from "@sveltejs/kit";
import { arktype } from 'sveltekit-superforms/adapters';
import { env } from "$env/dynamic/private"

import {
    productSchema,
    productDefaults,
    deleteProductSchema,
    deleteProductDefaults,
    categorySchema,
    categoryDefaults,
} from './schemas'

// PERF: Currently working on

export async function createCategory({ request, fetch }: RequestEvent) {
    const formValidated = await superValidate(request, arktype(categorySchema, { defaults: categoryDefaults }))
    if (!formValidated.valid) {
        if (formValidated.errors.name) return setError(formValidated, "name", "Le nom saisi n'est pas valide.")
        if (formValidated.errors.description) return setError(formValidated, "description", "La description est requise.")
        return setError(formValidated, "Le formulaire n'est pas valide. Veuillez vérifier les champs saisis.")
    }
    if (!formValidated.data.name.trim()) return setError(formValidated, "name", "Le nom est requis.")

    // The backend takes JSON, and only the name and description: a new
    // category is a draft, the chosen status is applied right after.
    const res = await fetch(`${env.API_URL}/admin/categories`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name: formValidated.data.name.trim(), description: formValidated.data.description }),
    })
    if (res.ok && formValidated.data.status !== "draft") {
        const { id } = await res.json()
        const statusRes = await fetch(`${env.API_URL}/admin/categories/${id}`, {
            method: "PATCH",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ status: formValidated.data.status }),
        })
        if (!statusRes.ok) {
            return setError(formValidated, "La catégorie a été créée en brouillon, mais son statut n'a pas pu être modifié.")
        }
    }
    if (!res.ok) {
        switch (res.status) {
            case 403:
                return setError(formValidated, "Vous n'avez pas les droits pour effectuer cette action.")
            case 404:
                return setError(formValidated, "La ressource demandée est introuvable.")
            case 409:
                return setError(formValidated, "Une catégorie porte déjà ce nom.")
            case 500:
                return setError(formValidated, "Une erreur serveur est survenue. Le service semble momentanément indisponible.")
            case 400:
                const errorMsg = await res.json()
                return setError(formValidated, "La requête que vous avez soumise contient des données incorrectes ou incomplètes. Veuillez vérifier les champs saisis et réessayer.", errorMsg.msg);
            default:
                return setError(formValidated, "Une erreur est survenue. Veuillez réessayer.");
        }
    }
    return {
        form: formValidated,
        success: true,
    }
}

/** The message shown when a category request fails. */
function categoryError(status: number): string {
    switch (status) {
        case 403:
            return "Vous n'avez pas les droits pour effectuer cette action."
        case 404:
            return "Cette catégorie n'existe plus."
        case 409:
            return "Une catégorie porte déjà ce nom, ou elle contient encore des produits."
        default:
            return "Une erreur est survenue. Veuillez réessayer."
    }
}

// The edit and delete forms of a category card: plain forms, each answered
// with { categoryError } on failure, so the card shows it in a toast.
export async function updateCategory({ request, fetch }: RequestEvent) {
    const data = await request.formData()
    const id = String(data.get("id") ?? "")
    const name = String(data.get("name") ?? "").trim()
    const status = String(data.get("status") ?? "")
    if (!id) return fail(400, { categoryError: "Catégorie introuvable." })
    if (!name) return fail(400, { categoryError: "Le nom est requis." })
    if (!["draft", "published", "archived"].includes(status)) return fail(400, { categoryError: "Statut invalide." })

    const res = await fetch(`${env.API_URL}/admin/categories/${id}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name, description: String(data.get("description") ?? ""), status }),
    })
    if (!res.ok) return fail(res.status, { categoryError: categoryError(res.status) })
    return { categoryUpdated: true }
}

export async function deleteCategory({ request, fetch }: RequestEvent) {
    const id = String((await request.formData()).get("id") ?? "")
    if (!id) return fail(400, { categoryError: "Catégorie introuvable." })

    const res = await fetch(`${env.API_URL}/admin/categories/${id}`, { method: "DELETE" })
    if (res.status === 409) {
        return fail(409, { categoryError: "Cette catégorie contient encore des produits : supprimez-les ou déplacez-les d'abord." })
    }
    if (!res.ok) return fail(res.status, { categoryError: categoryError(res.status) })
    return { categoryDeleted: true }
}

// NOTE: Done
export async function createProduct({ request, fetch }: RequestEvent) {
    const formData = await request.formData()

    const formValidated = await superValidate(formData, arktype(productSchema, { defaults: productDefaults }))

    if (!formValidated.valid) {
        if (formValidated.errors.name) return setError(formValidated, "name", "Le nom saisi n'est pas valide.")
        if (formValidated.errors.price) return setError(formValidated, "price", "Le prix saisi n'est pas valide.")
        if (formValidated.errors.category) return setError(formValidated, "category", "La catégorie est requise.")
        if (formValidated.errors.description) return setError(formValidated, "description", "La description est requise.")
        if (formValidated.errors.duration) return setError(formValidated, "duration", "La durée doit être un nombre valide.")
        if (formValidated.errors.bufferTime) return setError(formValidated, "bufferTime", "Le temps tampon doit être supérieur à 0.")
        if (formValidated.errors.cancellationHours) return setError(formValidated, "cancellationHours", "Le délai d'annulation doit être supérieur à 0.")
    }

    // image validation
    const image = formData.get("image")
    if (!image && formData.get("published") === "published") {
        console.log("the product should be published but there is no image")
        return setError(formValidated, "published", "Le status ne peut pas etre valide sans image publiee.")
    }
    if (image && image instanceof File && image.size > 0) {
        console.log("here this is the case where I check the type of image to see if it is a file")
        console.log("the image size:", image.size)
        console.log("the image type:", image.type)
        if (!image.type.startsWith('image/')) {
            // NOTE: this is where I went for some reason
            return fail(400, {
                form: formValidated,
                message: 'Uploaded file must be an image.'
            });
        }
    }

    // remove unecessary fields
    formData.delete("id")
    formData.delete("updatedAt")

    const res = await fetch(`${env.API_URL}/admin/products`, {
        method: "POST",
        body: formData,
    })

    if (!res.ok) {
        console.log("Something went wrong with the API response, status:", res.status)
        switch (res.status) {
            case 403:
                return setError(formValidated, "Vous n'avez pas les droits pour effectuer cette action.")
            case 404:
                return setError(formValidated, "La ressource demandée est introuvable.")
            case 409:
                return setError(formValidated, "Cette ressource est déjà utilisée et ne peut pas être supprimée.")
            case 500:
                return setError(formValidated, "Une erreur serveur est survenue. Le service semble momentanément indisponible.")
            case 400:
                const errorMsg = await res.json()
                return setError(formValidated, "La requête que vous avez soumise contient des données incorrectes ou incomplètes. Veuillez vérifier les champs saisis et réessayer.", errorMsg);
            default:
                return setError(formValidated, "Une erreur est survenue. Veuillez réessayer.");
        }
    }
    console.log("Product successfully created.")
    return {
        form: formValidated,
        success: true,
    }
}

export async function deleteProduct({ request, fetch }: RequestEvent) {
    const form = await superValidate(request, arktype(deleteProductSchema, { defaults: deleteProductDefaults }))
    console.log("here in the delete action")
    if (!form.valid) return setError(form, "id", "L'ID saisie n'est pas valide. Veuillez vérifier et réessayer.")

    const res = await fetch(`${env.API_URL}/admin/products/${form.data.id}`, {
        method: "DELETE",
    })

    if (!res.ok) {
        switch (res.status) {
            case 403:
                return setError(form, "Vous n'avez pas les droits pour effectuer cette action.")
            case 404:
                return setError(form, "La ressource demandée est introuvable.")
            case 409:
                return setError(form, "Cette ressource est déjà utilisée et ne peut pas être supprimée.")
            case 500:
                return setError(form, "Une erreur serveur est survenue. Le service semble momentanément indisponible.")
            case 400:
                const errorMsg = await res.json()
                return setError(form, "La requête que vous avez soumise contient des données incorrectes ou incomplètes. Veuillez vérifier les champs saisis et réessayer.", errorMsg.msg);
            default:
                return setError(form, "Une erreur est survenue. Veuillez réessayer.");
        }
    }
    return {
        form,
    }
}

export async function updateProduct() {
    console.log("here is the udpate action")
}

// Archiving a product (and bringing it back as a draft): it is never
// deleted, the bookings that refer to it keep it, but an archived product
// is no longer offered.
export async function setProductStatus({ request, fetch }: RequestEvent) {
    const data = await request.formData()
    const id = String(data.get("id") ?? "")
    const status = String(data.get("status") ?? "")
    if (!id) return fail(400, { productError: "Produit introuvable." })
    if (!["draft", "published", "archived"].includes(status)) return fail(400, { productError: "Statut invalide." })

    const res = await fetch(`${env.API_URL}/admin/products/${id}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ publishedStatus: status }),
    })
    if (!res.ok) {
        console.error("Product status update failed:", res.status, await res.text())
        return fail(res.status, { productError: "Le statut du produit n'a pas pu être modifié." })
    }
    return { productStatus: status }
}
