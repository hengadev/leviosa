// TODO: Finish that thing with the right text/option to setup the project

import {
    Accordion,
    AccordionContent,
    AccordionItem,
    AccordionTrigger,
} from "@/components/ui/accordion";

export default function Concept(){
    return (
        <Accordion type="single" collapsible>
            <AccordionItem value="item-1">
                <AccordionTrigger>
                    Presentation du concept ?
                </AccordionTrigger>
                <AccordionContent>
                    Le concept est le suvant : blablabla...
                </AccordionContent>
            </AccordionItem>

            <AccordionItem value="item-2">
                <AccordionTrigger>
                    Presentation du concept ?
                </AccordionTrigger>
                <AccordionContent>
                    Le concept est le suvant : blablabla...
                </AccordionContent>
            </AccordionItem>
            <AccordionItem value="item-3">
                <AccordionTrigger>
                    Presentation du concept ?
                </AccordionTrigger>
                <AccordionContent>
                    Le concept est le suvant : blablabla...
                </AccordionContent>
            </AccordionItem>
            <AccordionItem value="item-4">
                <AccordionTrigger>
                    Presentation du concept ?
                </AccordionTrigger>
                <AccordionContent>
                    Le concept est le suvant : blablabla...
                </AccordionContent>
            </AccordionItem>
        </Accordion>
    )
}
