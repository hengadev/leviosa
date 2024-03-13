import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import {useState} from "react"

import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
// import { Calendar } from "@/components/ui/calendar";

import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form"
import { Button } from "@/components/ui/button"

const monthsCount = 4;
const date = new Date();
const mois = [
    "Janvier",
    "Fevrier",
    "Mars",
    "Avril",
    "Mai",
    "Juin",
    "Juillet",
    "Aout",
    "Septembre",
    "Octobre",
    "Novembre",
    "Decembre",
];

let nextMonths = mois.slice(date.getMonth(), monthsCount);
if (nextMonths.length < monthsCount) {
    nextMonths = nextMonths.concat(mois.slice(0, monthsCount - nextMonths.length));
}
enum Month {
    Janvier = "Janvier",
    Fevrier = "Fevrier",
    Mars = "Mars",
    Avril = "Avril",
    Mai = "Mai",
    Juin = "Juin",
    Juillet = "Juillet",
    Aout = "Aout",
    Septembre = "Septembre",
    Octobre = "Octobre",
    Novembre = "Novembre",
    Decembre = "Decembre",
}

function onSubmit(values: z.infer<typeof formSchema>) {
    console.log("Je send mon vote", values);
}

// validation with zod
const formSchema = z.object({
    month: z.nativeEnum(Month),
    // week: z.number().min(1).max(4),
});

export default function EventForm() {
    const form = useForm<z.infer<typeof formSchema>>({
        resolver: zodResolver(formSchema),
        // TODO: Make the current next month
        defaultValues: {
            // month: Month.Janvier,
            month: Month[mois[(new Date().getMonth()+1)]],
            // week: 1,
        }
    })

    const [date, setDate] = useState<Date | undefined>(new Date())

    // TODO : The calendar is for admins to picks dates to add, I get that from the database
    // NOTE: The calendar that goes under the Form field
    // <Calendar
    //     mode="single"
    //     selected={date}
    //     onSelect={setDate}
    //     className="col-span-2 rounded-md border"
    // />
    
    return (

    <div>
    <Form {...form}>

    <FormField
        control={form.control}
        name="month"
        render={({ field }) => (
            <FormItem className="col-span-2 space-y-4 mb-4">
                <FormLabel>Mois</FormLabel>
                <Select
                    onValueChange={field.onChange}
                    defaultValue={field.value}
                >
                    <FormControl>
                        <SelectTrigger>
                            <SelectValue placeholder="Mois" />
                        </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                        {nextMonths.map((month) => (
                            <SelectItem key={month} value={month}>{month}</SelectItem>
                        ))}
                    </SelectContent>
                </Select>
                <FormMessage />
            </FormItem>
        )}
    />

    <FormField
        control={form.control}
        name="month"
        render={({ field }) => (
            <FormItem className="space-y-8 col-span-2">
                <FormLabel>
                    Choisis le jour qui t'interesse pour le mois de Fevrier (un
                    seul jour possible)
                </FormLabel>
                <FormControl>
                    <RadioGroup
                        onValueChange={field.onChange}
                        defaultValue={field.value}
                        // className="flex flex-col space-y-1"
                        className="grid w-[50%] space-y-4 mx-auto"
                    >
                        <FormItem className="bg-[#0f172a] text-white px-4 py-3 rounded-md flex items-center space-x-3 space-y-0">
                            <FormControl>
                                <RadioGroupItem value="first" />
                            </FormControl>
                            <FormLabel className="font-normal">
                                Samedi 7
                            </FormLabel>
                        </FormItem>
                        <FormItem className="bg-[#0f172a] text-white px-4 py-3 rounded-md flex items-center space-x-3 space-y-0">
                            <FormControl>
                                <RadioGroupItem value="second" />
                            </FormControl>
                            <FormLabel className="font-normal">
                                Samedi 14
                            </FormLabel>
                        </FormItem>
                        <FormItem className="bg-[#0f172a] text-white px-4 py-3 rounded-md flex items-center space-x-3 space-y-0">
                            <FormControl>
                                <RadioGroupItem value="third" />
                            </FormControl>
                            <FormLabel className="font-normal">
                                Samedi 21
                            </FormLabel>
                        </FormItem>
                        <FormItem className="bg-[#0f172a] text-white px-4 py-3 rounded-md flex items-center space-x-3 space-y-0">
                            <FormControl>
                                <RadioGroupItem value="fourth" />
                            </FormControl>
                            <FormLabel className="font-normal">
                                Samedi 28
                            </FormLabel>
                        </FormItem>
                    </RadioGroup>
                </FormControl>
                <FormMessage />
            </FormItem>
        )}
    />
    <div className="text-center">
        <Button className="w-[65%] text-lg mt-20" type="submit">Valider votre choix</Button>
    </div>
        </Form>
        </div>
    )
}
