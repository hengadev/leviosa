### Ce que je dois faire
- [x] Implement the S3 store for the images
- [x] Do the functionnality to send emails to some users with custom templates depending on the subject of the email
- [ ] Comment gerer les paiements (stripe ?)
- [ ] Commment gerer le fait que des paiements soient a realiser pour que les votes soient effectifs
    - [ ] Ajouter une date limite d'acceptation des paiements pour chaque event (field or something calculated when parsing the event with the cron job)
    - [ ] Tout vote non paye se voit annuler a la premiere data limite d'acceptation des paiements
    - [ ] Dans la table votes je dois ajouter un field hasPaid qui est un bool pour check si un user a paye pour l'event avant l'heure
- [ ] Do the cron jobs to send the email depending on some date that I get from the events for reminders for all the users
- [x] Il me faut un admin/mail pour que Livio puisse gerer l'envoi de mail en cas de pepin.
- [ ] Refacto ou je fais attention d'utiliser le plus possible des references pour speed mon code en fait.
- [ ] Refacto ou je design mieux mes fonctions, si on regarde les fonctions hasSession et parseCreatedAt, elles ne font pas exactement ce que leur nom suggere.

<!-- TODO: -->
<!-- Learn how to validate the user ie for a user value, go through each field and use some sort of validation to return if the field is fine. -->

### projets pour generer un qrcode a partir d'un URL : https://www.youtube.com/watch?v=OqPJ5SyowQg

### Prise de note

Est que le code que j'utilise ici me permet de faire du sql injection via les queries que je fais ?
Je pense que oui parce que je ne vois pas ou sont  les securites pour cela. 
Comment y remedie alors ?

401 status code => not authenticated to access a ressource

cookie set sameSite to same if you use the same URL for both the front and the backend and use the value none for other options
put that in secure (https), and http only (so no client accessing that value)

Pour curl si je veux utiliser des query parameters il faut mettre l'url entre ""
le flag -X est pour preciser la methode
curl -X GET "http://localhost:5000/event?id=1"

### Best practices to implement authorization in this [article](https://stackoverflow.blog/2021/10/06/best-practices-for-authentication-and-authorization-for-rest-apis/)
1. use tls (transport layer security)


#todo: store the session informations in a redis database just to see how it works ?

<!-- NOTE: Quand je sign up, je sign in ? -->

### Difference between using the body with json and the query strings
The body is to be used to send the information that needs to be downloaded or uploaded from the server
The query strings is to specify the data requested
-> Example : 
1. Je veux upload un ficher alors je mets dans le body son nom, type de fichier etc..
2. Si je veux recuperer une liste de fichier je peux par exemple precise dans les query strings que je veux tous les livres d'une certaine annee par exemple

Status 409 is the status that you can send back if  you try to post a ressource that already exists [from this web search](https://www.google.com/search?channel=fs&client=ubuntu-sn&q=what+status+to+send+if+I+tried+to+add+an+entry+to+the+database+that+already+exists)
It is equivalent to the http.StatusConflict (there is a cheet sheet to find the corresponding values)

bcrypt is an algorithm that implements a one way hashing function 


<!-- NOTE: Find what is a sync.Mutex and what is the usecase of that -->
In what context did I see this one ?

### Regarde au niveau de cette page sur les [headers](https://stackoverflow.com/questions/12830095/setting-http-headers) et ce qu'il y a set
Ils en parlent aussi [ici](https://stackoverflow.com/questions/28086582/return-code-for-wrong-http-method-in-rest-api)

### auth workflow (how to implement that because I want in a function to check for privileges) 
1. check si il y un cookie avec le nom sessionCookieName
    - si il n'y pas de cookie on envoie un message d'eerureur avec le status.. (unauthorized) puis je redirige vers la page sign in pour que quand la personne s'identifie elle puisse acceder a la page
2. je recupere le session_id de ce cookie
    - si session_id not in database then rediret to the sign in page
    - si session_id is in the database then get the username asscociated 
3. Use the username that you just got, to see if the user has the priviledge to do the action (use a function for that)
    - si c'est auth keep the action going, whatever that action is
    - else send an unauthorized status Code



### There is a difference between using var and not using it
With var value SomeType, we declare the variable but it is not initialised. So we have to initialise it later with value = ...
With the other one, the var is declared and intialised with an emputy slice

### Mes cron jobs
parcourir la liste des events, parse la date et ensuite si un certain delai a partir de la date est respectee envoie un mail a tous ceux inscrits.
Je vais faire cette tache tous les jours

cron job syntax
* * * * * 
minute - hour - day of moth - month of year - day of week

every day of the week at 5:30 utc
30 5 * * *

Il me faut un cron job pour les mails et pour gerer la base de donnee notamment la duplication que je vais sur git apres encrytion de toutes les donnees.

### How to connect to the database using the tim Pope plugin
sqlite://[filename]
Then give a name to the connection

### Organise the file a little bit better
En gros j'ai un probleme dans mon ficier api ou pour chaque fichier j'en ai un en _test associe et c'est illisible

### For the S3 storage
I get all the information from this [video](https://www.youtube.com/watch?v=ssLcGwHv7Hc)
- how to serve static files from a folder ?
- what is html rendering and how to do it with the std library
- use gin to copy some of their function 
    - how to get the file (using the FormFile function)
    - how to save the file (using the SaveUploadedFile function)

On retrouve les libs qu'il faut via la doc a ces liens :
Partie install the aws sdk for go v2
- https://aws.github.io/aws-sdk-go-v2/docs/getting-started/
Sur la premiere instance de code on installe ce qui nous manquait :
- https://aws.github.io/aws-sdk-go-v2/docs/sdk-utilities/s3/

### Account to make for the setup
- make an AWS account
- make a stripe account

### Projects for cybersecurity that I found on some tiktok but no need for the link.
create password manager 
create malware
create packet sniffer

msg := "content-length: 37 \r\n\r\n{\"jsonrpc\": \"2.0 -> form the video of teej devries where he was making a lsp in golang.

### Prise de note sur les cors et sur les csrf parce que je ne sais pas ce que c'est.
- same origin policy :
    - the orgin of an url is the combination of : scheme (protocol = http for example) / domain / port or scheme + authority 
    Exemple : De l'url : https://example.fr:443/foo -> the origin is https://example.fr:443
    the /foo part is part of the url but not the origin.
    The authority is the combination of the domain and the port used.  In the previous example the authority is example.fr:443

    the port indicates the technical gates used to acccesses the ressources on the web server.

    If we use the standard port of the http protocol (80 for http and 443 for the https) then it is omitted in the authority.

    <!-- NOTE: url that does not use an authority is mailto:foobar. It still uses a scheme but here no authority. In the previous example the : separeated the scheme from the rest of the url and the // was to indicate the beginning of the authority. -->

    the part /foo in the previous example is the path to ressources

    then we have parameter "?id=someid&name=somename"

- CORS (cross orign ressource shared) :
    - CORS is a header mechanism that allows you to specify any origin that your server may access.

Look for the preflight request, thaat has to do with CORS. Basically when the request can cause cors the client makes a preflight request to ensure that the option used in the header are available

### From this [video](https://www.youtube.com/watch?v=h2cVqKLcf2A) 
A 10:30, il parle de sa solution pour les paiements et il n'utilise pas stripe mais lemon squeezy mais je sais pas si je peux lier ca a golang.

### http status code
1xx informational
2xx success
3xx redirect
4xx client error
5xx server error

### Comment gerer le domain name :
- Deux choses, je veux le front et le backend :
1. Je peux faire quelque chose comme http://myapp.example/ pour le frontend et http://api.myapp.example/ pour le backend (subdomain).
2. something like   http://myapp.example/ et http://myapp.example/api
Les deux trucs seront deployes a partir d'un docker container.

NOTE: The thing is that the api is not to be accessed by third parties but only by the app itself

### I need to learn about this env variables
From this [article](https://blog.devgenius.io/why-a-env-7b4a79ba689) what does it mean to be in production or developper mode ?
En gros c'est pour gerer des variables pour setup le projet mais que l'on ne veut pas accessible a tout le monde (exemple autre partcipant quu vont devoir set leurs propres env variables de leur cote)

### I have seen people talking about pingora which is the cloudfare replacement for nginx
I want to learn about that and to use as much as possible the documentation to use it as a real developper would.
A few links to see that :
- https://www.youtube.com/results?search_query=pingora


### Comprendre le stripe API: 
Je viens de cette [page](https://docs.stripe.com/payments/checkout) !


