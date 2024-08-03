import { redirect } from "@sveltejs/kit"
import type { PageServerLoad } from "./$types"

type Mail = {
    eventId: string
    content: string,
}

type Notification = {
    eventId: string
    content: string,
}

export const load: PageServerLoad = async ({ locals }) => {
    if (!locals.user) {
        console.log("redirect")
        throw redirect(301, "/")
    }
    // NOTE: How I want to do it when everything setup.
    // const userId = locals.user.id
    // const res = await fetch("http://localhost:5000/api/v1/mails", {
    //     method: "POST",
    //     headers: { "Content-Type": "application/json" },
    //     body: JSON.stringify({ "userId": userId })
    // })
    // const mails = await res.json()
    const mails: Mail[] = [
        {
            eventId: "323fd3g82034f",
            content: "Je teste l'envoi de mail.",
        },
        {
            eventId: "4f338423gt27a",
            content: "Un autre test d'envoi de mail.",
        },
        {
            eventId: "aw349bvq34g9q",
            content: "Un autre test d'envoi de mail.",
        },
        {
            eventId: "34utq3g95g340",
            content: "Enfin on peut faire du code les gars.",
        },
        {
            eventId: "22f0avdv021e",
            content: "Ici, on envoie du texte pour simuler une histoire de mail recue par les gens.",
        },
    ]

    const notifications: Notification[] = [
        {
            eventId: "323fd3g82034f",
            content: "Je teste l'envoi de mail.",
        },
        {
            eventId: "4f338423gt27a",
            content: "Un autre test d'envoi de mail.",
        },
        {
            eventId: "4f338423gt27a",
            content: "Un autre test d'envoi de mail.",
        },
    ]
    return { mails, notifications }
}
