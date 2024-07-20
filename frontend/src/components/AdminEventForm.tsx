"use client"

import { zodResolver } from "@hookform/resolvers/zod"
import { format } from "date-fns"
import { formatRFC3339 } from "date-fns"
import { CalendarIcon } from "lucide-react"
import { useForm } from "react-hook-form"
import { z } from "zod"

import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import { Calendar } from "@/components/ui/calendar"
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
import {
    Popover,
    PopoverContent,
    PopoverTrigger,
} from "@/components/ui/popover"

const FormSchema = z.object({
    date: z.date({
        required_error: "A date for the event is required.",
    }),
    // TODO: remove required error because you want to be able to create an event withtout having the place set yet
    // TODO: Can I use google maps to get the exact address ?
    location: z.string({
        required_error: "A date for the event is required.",
    }),
    placecount: z.coerce.number({
        required_error: "A number of places need to be given for the event.",
    }).min(1)
})

// TODO: add field localisation, number of places

export default function AdminEventForm() {
    const form = useForm<z.infer<typeof FormSchema>>({
        resolver: zodResolver(FormSchema),
        defaultValues: {
            location: "",
            placecount: 0,
        }
    })

    async function postData(data: any) {
        try {
            const formatData = { ...data, date: formatRFC3339(data.date) }
            console.log("The values sent are:", JSON.stringify(data))
            console.log("The other values with the new formatting sent are:", JSON.stringify(formatData))
            // TODO: format the data using the rfc3339 standard

            const res = await fetch("http://localhost:5000/admin/events", {
                method: "POST",
                mode: "cors",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify(formatData),
            })

            if (!res.ok) throw new Error("Failed to get a response from server")
        }
        catch (err) {
            console.warn("Something went wrong : ", err)
        }
    }

    function onSubmit(data: z.infer<typeof FormSchema>) {
        // TODO: How to get the value from the data
        console.log("On submit ici, print data :", data)
        postData(data)
    }

    return (
        <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
                <FormField
                    control={form.control}
                    name="date"
                    render={({ field }) => (
                        <FormItem className="flex flex-col">
                            <FormLabel>Date de l'evenement</FormLabel>
                            <Popover>
                                <PopoverTrigger asChild>
                                    <FormControl>
                                        <Button
                                            variant={"outline"}
                                            className={cn(
                                                "w-[240px] pl-3 text-left font-normal",
                                                !field.value && "text-muted-foreground"
                                            )}
                                        >
                                            {field.value ? (
                                                format(field.value, "PPP")
                                            ) : (
                                                <span>Choisissez une date</span>
                                            )}
                                            <CalendarIcon className="ml-auto h-4 w-4 opacity-50" />
                                        </Button>
                                    </FormControl>
                                </PopoverTrigger>
                                <PopoverContent className="w-auto p-0" align="start">
                                    <Calendar
                                        mode="single"
                                        selected={field.value}
                                        onSelect={field.onChange}
                                        initialFocus
                                    />
                                </PopoverContent>
                            </Popover>
                            <FormDescription>La date de l'evenement est a fournir pour planifier celui-ci</FormDescription>
                            <FormMessage />
                        </FormItem>
                    )}
                />
                <div className="grid grid-cols-2 gap-8">
                    <FormField
                        control={form.control}
                        name="location" render={({ field }) => (<FormItem> <FormLabel>Location</FormLabel>
                            <FormControl>
                                <Input className="w-full px-4 py-6" placeholder="Addresse de l'evenenement" {...field} value={field.value ?? ''} />
                            </FormControl>
                            <FormMessage />
                        </FormItem>
                        )}
                    />
                    <FormField
                        control={form.control}
                        name="placecount"
                        render={({ field }) => (
                            <FormItem>
                                <FormLabel>Placecount</FormLabel>
                                <FormControl>
                                    <Input className="w-full px-4 py-6" placeholder="Nombre de place pour l'evenement" {...field} value={field.value ?? ''} />
                                </FormControl>
                                <FormMessage />
                            </FormItem>
                        )}
                    />
                    <Button className="mx-auto mt-16 py-6 text-xl col-span-2 w-[60%]" type="submit">Submit</Button>
                </div>
            </form>
        </Form>
    )
}
