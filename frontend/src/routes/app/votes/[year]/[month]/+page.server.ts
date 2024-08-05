import { redirect } from "@sveltejs/kit";
import type { Action } from "./$types.js";
import { API_URL } from "$env/static/private"

type Vote = {
    day: number,
    month: number,
    year: number,
}

// TODO: make sure to place that array somewhere else since I use it on the /app/votes too
const months = ["janvier", "fevrier", "mars", "avril", "mai", "juin", "juillet", "aout", "septembre", "octobre", "novembre", "decembre"]

export async function load({ locals, cookies, params }) {
    if (!locals.user) {
        throw redirect(302, "/")
    }
    const year = Number(params.year)
    const month = Number(months.indexOf(params.month)) + 1
    console.log("month", month)
    const sessionId = cookies.get("sessionId");
    const res = await fetch(`${API_URL}/api/v1/vote?year=${year}&month=${month}`, {
        method: "GET",
        headers: {
            "Authorization": `Bearer ${sessionId}`,
        },
    })
    const { isDefault, votes } = await res.json()
    return { isDefault, votes }
}

export const actions: Action = {
    default: async ({ request, cookies, params }) => {
        console.log("Sending the data to the backend !")
        const formData = await request.formData()
        const count = Number(formData.get("count"))
        const year = Number(params.year)
        const month = Number(months.indexOf(params.month))
        const votes: Vote[] = []
        const sessionId = cookies.get("sessionId");
        for (let i = 0; i < count; i++) {
            votes.push({
                day: Number(formData.get(String(i))),
                month,
                year,
            })
        }
        const body = JSON.stringify(votes)
        console.log("the body is : ", body)
        const res = await fetch(`${API_URL}/api/v1/votes`, {
            method: "POST",
            headers: { "Authorization": `Bearer ${sessionId}`, "Content-Type": "application/json", },
            body,
        })
        // TODO: Do something when successful.
        console.log("the status is : ", res.status)
        const data = await res.json()
        console.log("the data is : ", data)
        console.log("the message, I get is ", data.message)
        if (res.status === 201) {
            console.log("Done!")
        }
    }
}
