import { redirect } from "@sveltejs/kit"

// TODO: 
// - how many dates for the month (pick the x most ranked)
// - ranked the date by availabiilty
// - put an option I do not care (so you can choose to not ponder some dates.)

// TODO: ce que doit contenir le vote en question
// - Y a t il une location precisee. (si oui un certain affichage, sinon un autre)
// - Quel type d'evemement c'est ? Si plusieurs types existent 

export const actions = {
    default: async ({ request }) => {
        const formData = request.formData()
        const year = formData.get("year")
        const month = formData.get("month")
        throw redirect(302, `/app/votes/${year}/${month}`)
    }
}

export async function load({ locals, cookies }) {
    if (!locals.user) {
        throw redirect(302, "/")
    }
    const sessionId = cookies.get("sessionId");
    const res = await fetch("http://localhost:5000/votes", {
        method: "GET",
        headers: {
            "Authorization": `Bearer ${sessionId}`,
        },
    })
    // TODO: do the error handling in here.
    // if (!res.ok) {
    //     console.log("the status for the request is : ", res.status)
    //     const data = await res.json()
    //     console.log("the response object is : ", data)
    // }
    // console.log("the res is :", res)
    const { futurVotes } = await res.json()
    return { futurVotes }
}
