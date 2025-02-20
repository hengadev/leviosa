type Card = {
    date: string,
}

const cards: Card[] = [
    {
        date: "12/12/2024",
    }
]

type PageRes = { cards: Card[], role: import("$lib/types").Role }
export function load({ locals }): PageRes {
    return { cards, role: locals.user.role }
}
