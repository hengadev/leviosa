### Un cookie n'est valable que pendant une session du navigateur ?


<!-- TODO:  -->
<!-- 1. Do the animation for the different pages. -->

### The message from the successful request
Sending the data to the backend !
the body is :  [{"day":24,"month":4,"year":2024},{"day":23,"month":4,"year":2024},{"day":27,"month":4,"yea
r":2024},{"day":30,"month":4,"year":2024}]
the status is :  201
the data is :  { message: 'Vote created successfully' }
Done!


### The redis store I need to use and more generally how to organize my backend.
From the video : https://www.youtube.com/watch?v=50BYTzwC14Y&t=1218s
I need a redis store to store the session informations.
the keys are going to be the session id The value are going to be the firstname, lastname, id and email
From the golang part of the backend, he sends the session id as an authorization header to the sveltekit backend. Puis on set le cookie sur 
le sveltekit backend.

-> Des reponses que je trouve pertinentes sur la video de Ben Davis aussi :
> First : 
1. You can through a `use:enhance` action on form for progressive enhancement
2. In your sign in route action you should use the fetch provided by sveltekit. It is available on the event just like request, params and etc so you could do {request, fetch} in your default action to destructure event.
3. Akin to fetch you also have a setHeaders function that can be destructred from event as well

> Second :
You can expose your golang backend as a binary and execute directly inside your SvelteKit server, that way you eliminate the extra HTTP roundtrip. Your app will run like the older CGI servers


### The backend organisation 
the same that I have in my application but all the routes start with /api because I am going to have my frontend and 
my backend on the same server (so that I do not have any CORS policy)

#### The route for that page.
/ (for the signin page)
/signup (for the signup if it is a new user)
/app (one the user is signed in)
    - /app/events
        - /events/ (get) -> get all the events old and future for the specific user
        - /events/:eventId (get) -> the detail of the event for more explanation.
        - /events/:eventId/photos (get) -> get all the photos for a certain user
        - /events/:eventId/booking (post) -> make a booking for a certain event

    - /app/bookings (reservation done for specific events) 
        - /bookings/ (get) -> get all the bookings old and future for the specific user

    - /app/votes 
        -/votes/ -> get all the votes old and future for the specific user
        - /votes/:year/:month (get) -> make a vote for the specific month and year

    - /app/settings /settings/me (get, post) -> change the settings for a certain user

    - /app/checkout (simple message then redirect to main page) (use that video to get a feel of what needs to be done.)
        - /checkout/success
        - /checkout/cancel

    - /app/admin
<!-- TODO: get all the routes here so that I can make the modification and that routes need to be protected so that no one can access this -->

### Organise the colors for the project
// NOTE: old bg color blue :#0f172a
// color gris; rgba(60,60,67)
// gris metallique texte: #3c3c43
// light grey bg : #ebebef
// la couleur des subtitles sur le bg body : rgba(60, 60, 67, 0.78)

### Voir le produit shipfast de Marc Lou pour voir les outils qu'il utilise.
- Il utilse un email provider (bonne idee ou je le fais moi meme ?)
- Il utilise Auth js mais il ne precise utiliser que google login (oauth surement) et des magic link.

### Des liens d'exemple pour le design que je peux me faire : 
- https://dribbble.com/shots/19368315-Eventioe-Event-Booking-App-UI-Kit
- https://dribbble.com/shots/20111667-Event-Booking-App/attachments/15189261?mode=media
- https://cdn.dribbble.com/userupload/4613493/file/original-fd79a5e39acc30394da06f390aee5d3f.png?resize=752x
- https://cdn.dribbble.com/userupload/13023438/file/original-894ec57ca1e2565a8d8c5741a3f2b199.jpg?resize=752x
- https://cdn.dribbble.com/userupload/10156262/file/original-c3a04492b24aa866df2d2c8f59dac51d.jpeg?resize=752x
- https://dribbble.com/shots/19359042-Design-for-Recruitment-Progressive-Web-Application
- https://dribbble.com/shots/17608286-OAE-Progressive-Web-App-for-Orchestra-Event-Program

### The link I used for the tab section since I like it !
https://dribbble.com/shots/23248031-Dark-UI-for-a-notifications-modal

### Solution I used to persist the store on local storage.
https://stackoverflow.com/questions/56488202/how-to-persist-svelte-store

### Des petits sites que j'apprecies en terme d'esthetique : 
https://dribbble.com/shots/21560559-Governer-Election-App-Polling-App
https://dribbble.com/shots/18241310-Platform-for-holding-public-voting
https://dribbble.com/shots/16625925-Voting-Application
https://dribbble.com/shots/21657856-Voting-App
