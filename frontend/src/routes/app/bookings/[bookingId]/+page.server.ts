import type { PageServerLoad } from "./$types"

async function getBookingInformation(bookingId: string) {
    // const bookings = ["booking1", "booking2", "booking3", "booking4"]
    const res = await fetch(`http://localhost:5000/bookings?bookingId=${bookingId}`, {
        method: "GET",
    })
    const bookings = await res.json()
    // console.log(res)
    return bookings
}

// export const load: PageServerLoad = () => {
//     return {
//         getBookingInformation
//     }
// }
