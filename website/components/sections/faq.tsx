"use client";

import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "@/components/ui/accordion";
import { BlurFade } from "@/components/ui/blur-fade";
import { SectionWrapper, SectionHeading } from "@/components/common";
import { FAQ_ITEMS } from "@/lib/constants";

export function FAQ() {
  return (
    <SectionWrapper id="faq">
      <SectionHeading title="Frequently Asked Questions" />
      <BlurFade delay={0.2} inView>
        <div className="mx-auto max-w-3xl">
          <Accordion type="single" collapsible className="space-y-2">
            {FAQ_ITEMS.map((item, i) => (
              <AccordionItem
                key={i}
                value={`item-${i}`}
                className="border-border-subtle rounded-lg bg-surface-light/30 px-6 data-[state=open]:bg-surface-light/50"
              >
                <AccordionTrigger className="text-left text-text-primary hover:text-brand hover:no-underline py-5">
                  {item.question}
                </AccordionTrigger>
                <AccordionContent className="text-text-secondary pb-5 leading-relaxed">
                  {item.answer}
                </AccordionContent>
              </AccordionItem>
            ))}
          </Accordion>
        </div>
      </BlurFade>
    </SectionWrapper>
  );
}
