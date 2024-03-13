"use client"

import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import * as z from "zod";

import { Button } from "@/components/ui/button"
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form"
import { Input } from "@/components/ui/input"

const formSchema = z.object({
    email: z.string({
        required_error: "Cette adresse mail n'est pas valide. Format attendu johndoe@email.com",
    }).email(),
    // TODO: Change that to have more explicit validation from zod
    password: z.string({
        required_error: "Veuillez entrer votre addresse",
    }).min(8).max(20),
})

async function postData(url: URL | string, data = {}) {
  const response = await fetch(url, {
    method: "POST", 
    mode: "cors", 
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(data), 
  });
  return response;
}

// TODO: Make that function so that I can send it to my app.
async function onSubmit(values: z.infer<typeof formSchema>){
    // const data = {email: "admin@example.fr", password: "adminpassword"}
    // const data = {email: "admin|example.fr", password: "adminpassword"}
    const data = JSON.stringify(values)
    try {
        // TODO: Use a HOST env variable for that
        const url = new URL("http://localhost:5000/signin")
        const res = await postData(url, data)
        if (!res.ok) throw new Error("Failed to get a response from the server.")
        switch (res.status) {
            case 200: // OK
                // TODO: go to the home page using a router or something.
                // Astro.redirect('/')
                console.log("the user is authenticated my friend");
                break;
            case 401: // UNATHORIZED
                // les identifiants sont corrects mais n'ont pas de correspondance dans la db
                // TODO: This user does not exists UI to be printed
                console.log("user is not registered or email or password are wrong");
                break;
            case 403: // FORBIDDEN
                // les identifiant ne peuvent corresponde a quelque chose correct
                // TODO: The email or the password are not in the correct format UI to be printed
                console.log("email or password are not in the correct format.");
                break;
            default: 
                // console.log("got another statuscode not expected")
                throw new Error("Status code returned is not one that should be expected.")
        }
    } catch(err: any) {
        console.warn("There is an error : ", err.message)
    }
}

// TODO: Try to use that function to retrieve something from the S3 store that I set up ?
// TODO: Try to use that function to test the stripe api from the frontend.
async function handleClick() {
    try{
    // NOTE: old, i thought I needed the event id to get the price but it not the case.
    // const res = await fetch("http://localhost:5000/checkout?event_id=123", {
    const res = await fetch("http://localhost:5000/checkout", {
        method: "GET",
        mode: "cors", 
    })
    if (!res.ok) throw new Error("Failed to get a response from the server.")
    switch (res.status) {
        case 200:
        console.log("Successful request !")
        console.log(res.url)
        break;
        default: 
        console.log("Je suis sur de recuperer un status ici")
    }
    } catch(err) {
        console.warn("Something went wrong : ", err)
    }
}
// const val = https://checkout.stripe.com/c/pay/cs_test_a1GSjN1XbNxXDTIj05P2UeXiL1if2QimE9RoaqWKlzOjfxmVG0qAEw6MI4


export default function ProfileForm() {
    const form = useForm<z.infer<typeof formSchema>>({
        resolver: zodResolver(formSchema),
        defaultValues: {
            email: "",
            password: "",
        }
    })

  return (
    <div className="space-y-8"> 
        <div className="py-4 px-8 border-2 rounded-lg">
                <button className="border-2 px-4 py-2" onClick={handleClick}>Test</button>
            <Form {...form}>
              <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
                <div className="grid my-auto gap-4 space-y-4">
                    <h2 className="col-span-2 font-black text-2xl">Register</h2> 
                    <div className="col-span-2">
                        <FormField 
                            control={form.control}
                            name="email"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>Email</FormLabel>
                                    <FormControl>
                                        <Input 
                                            placeholder="Votre adresse email" {...field} 
                                            className="p-6"    
                                        />
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />
                    </div>
                    <div className="col-span-2">
                        <FormField
                          control={form.control}
                          name="password"
                          render={({ field }) => (
                            <FormItem>
                              <FormLabel>Mot de passe</FormLabel>
                              <FormControl>
                                <Input 
                                    placeholder="Votre mot de passe" {...field} 
                                    className="p-6"    
                                    type="password"
                                />
                              </FormControl>
                              <FormMessage />
                            </FormItem>
                          )}
                        />
                    </div>
                    <Button className="uppercase text-lg mt-20 col-span-2" type="submit">se connecter</Button>
                    <a href="" className="col-span-2 text-center hover:underline hover:underline-offset-4                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    ">Mot de passe oublie ?</a>
                </div>
              </form>
            </Form>
        </div>
        <div className="text-center mt-auto py-4 px-8 border-2 rounded-lg">
            <p>Nouveau sur notre plateforme ?</p>
            <Button className="uppercase w-full mt-4 text-xl">
                <a href="/register" className="">S'inscrire</a>
            </Button>
        </div>
  </div>
  )
}
