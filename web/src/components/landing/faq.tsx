import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "~/components/ui/accordion";

export default function FAQ() {
  return (
    <section className="py-24 bg-slate-50 dark:bg-slate-900/50">
      <div className="container px-4 md:px-6 max-w-3xl">
        <div className="text-center mb-16">
          <h2 className="text-3xl font-bold tracking-tight text-slate-900 dark:text-slate-50 sm:text-4xl">
            Frequently Asked Questions
          </h2>
          <p className="mt-4 text-lg text-slate-500 dark:text-slate-400">
            Everything you need to know about NexusOS.
          </p>
        </div>

        <Accordion type="single" collapsible className="w-full">
          <AccordionItem value="item-1">
            <AccordionTrigger>Is it Next.js 15 or 16?</AccordionTrigger>
            <AccordionContent>
              We are using the latest stable Next.js 16 with App Router and React 19 support.
            </AccordionContent>
          </AccordionItem>
          <AccordionItem value="item-2">
            <AccordionTrigger>What is "Fluid Density"?</AccordionTrigger>
            <AccordionContent>
              Fluid Density is our proprietary system that allows the UI to switch between a spacious "Comfort" mode (ideal for SaaS) and a dense "Compact" mode (ideal for Enterprise data entry) instantly.
            </AccordionContent>
          </AccordionItem>
          <AccordionItem value="item-3">
            <AccordionTrigger>Can I use this with Laravel or Django?</AccordionTrigger>
            <AccordionContent>
              Yes! The Enterprise license includes a pure Vite/React version that is easy to integrate into legacy backends like Laravel, Django, or ASP.NET.
            </AccordionContent>
          </AccordionItem>
          <AccordionItem value="item-4">
            <AccordionTrigger>Do you offer a refund?</AccordionTrigger>
            <AccordionContent>
              We offer a 30-day money-back guarantee if you find a critical bug that we cannot fix.
            </AccordionContent>
          </AccordionItem>
        </Accordion>
      </div>
    </section>
  );
}