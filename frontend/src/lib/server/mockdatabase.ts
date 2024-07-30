// Just a file to mock a database in fact. 

type User = {
    id: string,
    firstname: string,
    lastname: string,
    email: string,
    password: string,
}

const makeId = () => "id" + Math.random().toString(16).slice(2)

const users: User[] = [
    {
        id: makeId(),
        firstname: "Livio",
        lastname: "HENRY",
        email: "henrylivio@hotmail.com",
        password: "123soleil",
    },
    {
        id: makeId(),
        firstname: "Gary",
        lastname: "HENRY",
        email: "henry.gary@hotmail.com",
        password: "123soleil",
    },
    {
        id: makeId(),
        firstname: "Chantal",
        lastname: "HENRY",
        email: "henry.chantal@hotmail.com",
        password: "123soleil",
    },
    {
        id: makeId(),
        firstname: "Serge",
        lastname: "HENRY",
        email: "henryserge@hotmail.com",
        password: "123soleil",
    },
]

export function getAllUsers() {
    return users
}

type Event = {
    id: number,
    location: string,
    date: string,
}

const events: Event[] = [
    {
        id: Date.now(),
        location: "Some location",
        date: "02/12/2024",
    },
]

export function getAllEvents() {
    return events
}
