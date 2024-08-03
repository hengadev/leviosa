import type { PageLoad } from "./$types"

async function handleCheckout(eventId: string) {
    try {
        // TODO: send the event information through the body because more secure ?
        // const body = JSON.stringify(eventId)
        const res = await fetch(`http://localhost:5000/api/v1/checkout?eventId=${eventId}`, {
            method: 'POST',
            mode: 'cors',
            // body
        });
        const message = await res.json()
        // if (!res.ok) throw new Error('Failed to get a response from the server.');
        // NOTE: Replace allow to not have that page in the history so that is something to keep if I do not want the user to go back.
        // window.location.replace(message.url)
        // NOTE: Alternative si le fait d'avoir la page dans l'historique ne me derange pas
        window.location.href = message.url;
    } catch (err) {
        console.warn('Something went wrong : ', err);
    }
}

export const load: PageLoad = ({ data }) => {
    const { event } = data
    return { handleCheckout, event }
}
