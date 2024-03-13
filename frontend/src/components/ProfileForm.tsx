"use client"

import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import * as z from "zod";

import { Button } from "@/components/ui/button"
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form"
import { Input } from "@/components/ui/input"


// zod for validating the form
const formSchema = z.object({
    firstname: z.string({
        required_error: "Veuillez entrer votre prenom",
    }).min(2, {message: "Ce prenom est trop court."}).max(50, {message: "Ce message est trop long."}),
    lastname: z.string({
        required_error: "Veuillez entrer votre nom",
    }).min(2).max(50),
    // TODO: validate with a regex.
    telephone: z.string({
        required_error: "Ce numero de telephone n'est pas valide. Format attendu 0612345678",
    }).min(10).max(10),
    email: z.string({
        required_error: "Cette adresse mail n'est pas valide. Format attendu johndoe@email.com",
    }).email(),
    password: z.string({
        required_error: "Veuillez entrer un mot de passe", 
    }).min(10).max(50),
    adresse1: z.string({
        required_error: "Veuillez entrer votre addresse",
    }).min(2).max(50),
    adresse2: z.string({
        required_error: "Veuillez entrer votre addresse",
    }).min(2).max(50),
    ville: z.string({
        required_error: "Veuillez entrer votre ville",
    }).min(2).max(50),
    codepostal: z.number({
        required_error: "Veuillez entrer votre code postal",
    }).min(5).max(5),
    // TODO: Add the password field for the account 
})

function onSubmit(values: z.infer<typeof formSchema>){
    console.log("The value of the form are : ", values);
}

export default function ProfileForm() {
    const form = useForm<z.infer<typeof formSchema>>({
        resolver: zodResolver(formSchema),
        defaultValues: {
            firstname: "",
            lastname: "",
            telephone: "",
            email: "",
            password: "",
            adresse1: "",
            adresse2: "",
            ville: "",
            codepostal: 0,
        }
    })

  return (
    <div className="p-4 border-2 rounded-lg">
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
        <div className="grid grid-cols-2 gap-4 space-y-4">
        <h2 className="col-span-2 font-black text-2xl">Information de base<span className="text-red-400">*</span></h2>
        <FormField
          control={form.control}
          name="firstname"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Prenom</FormLabel>
              <FormControl>
                <Input placeholder="Votre prenom" {...field} />
              </FormControl>
              <FormDescription>
                Ceci est votre prenom.
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="lastname"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Nom</FormLabel>
              <FormControl>
                <Input placeholder="Votre nom" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="email"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Email</FormLabel>
              <FormControl>
                <Input placeholder="Votre adresse email" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="telephone"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Telephone</FormLabel>
              <FormControl>
                <Input placeholder="Votre numero de telephone" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <h2 className="col-span-2 font-black text-2xl">Adresse <span className="text-red-400">*</span></h2>
        <div className="col-span-2">
            <FormField
              control={form.control}
              name="adresse1"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Adresse</FormLabel>
                  <FormControl>
                    <Input placeholder="Votre adresse" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
        </div>
        <div className="col-span-2">
            <FormField
              control={form.control}
              name="adresse2"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Adresse (Ligne complementaire)</FormLabel>
                  <FormControl>
                    <Input placeholder="Votre adresse" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
        </div>
        <FormField
          control={form.control}
          name="ville"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Ville</FormLabel>
              <FormControl>
                <Input placeholder="Votre ville" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="codepostal"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Code Postal</FormLabel>
              <FormControl>
                <Input placeholder="Votre codePostal" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <Button className="text-lg col-span-2" type="submit">Enregistrer votre profile</Button>
        </div>
      </form>
    </Form>
        </div>
  )
}


// {getNextMonths.map(( month, id ) => (
//     <SelectItem key={id} value={month.toString()}>{month}</SelectItem>
// ))}
 

