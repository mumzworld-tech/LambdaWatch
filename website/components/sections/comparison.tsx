"use client";

import { CircleCheck, CircleX, Zap } from "lucide-react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { ScrollArea, ScrollBar } from "@/components/ui/scroll-area";
import { BlurFade } from "@/components/ui/blur-fade";
import { ShineBorder } from "@/components/ui/shine-border";
import {
  SectionWrapper,
  SectionHeading,
  GlassmorphicCard,
} from "@/components/common";
import { COMPARISON_FEATURES, COMPARISON_PRODUCTS } from "@/lib/constants";
import { cn } from "@/lib/utils";

function CellValue({ value }: { value: boolean | string }) {
  if (typeof value === "string") {
    return (
      <span className="text-sm md:text-base font-medium text-text-secondary">
        {value}
      </span>
    );
  }
  return value ? (
    <CircleCheck className="h-5 w-5 md:h-6 md:w-6 text-brand-green" />
  ) : (
    <CircleX className="h-5 w-5 md:h-6 md:w-6 text-text-muted/50" />
  );
}

export function Comparison() {
  return (
    <SectionWrapper id="comparison">
      <SectionHeading
        title="How We Compare"
        subtitle="See how LambdaWatch stacks up against alternatives."
      />
      <BlurFade delay={0.2} inView>
        <div className="max-w-5xl mx-auto">
          <GlassmorphicCard className="p-0 overflow-hidden">
            {/* Mobile: scrollable table */}
            <div className="md:hidden">
              <ScrollArea className="w-full">
                <Table>
                  <TableHeader>
                    <TableRow className="border-border-subtle hover:bg-transparent">
                      <TableHead className="w-[200px] pl-6 text-sm uppercase tracking-wider font-semibold text-text-muted">
                        Feature
                      </TableHead>
                      {COMPARISON_PRODUCTS.map((product) => (
                        <TableHead
                          key={product.name}
                          className={cn(
                            "text-center min-w-[120px] relative",
                            product.highlighted
                              ? "text-brand font-semibold"
                              : "text-text-muted"
                          )}
                        >
                          <div className="flex items-center justify-center gap-1.5">
                            {product.highlighted && (
                              <Zap className="h-4 w-4" />
                            )}
                            {product.name}
                          </div>
                          {product.highlighted && (
                            <ShineBorder
                              shineColor={["#FF9900", "#FFB84D"]}
                              borderWidth={1}
                            />
                          )}
                        </TableHead>
                      ))}
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {COMPARISON_FEATURES.map((feature, rowIdx) => (
                      <TableRow
                        key={feature}
                        className="border-border-subtle hover:bg-surface-light/30"
                      >
                        <TableCell className="pl-6 font-medium text-text-secondary">
                          {feature}
                        </TableCell>
                        {COMPARISON_PRODUCTS.map((product) => (
                          <TableCell
                            key={product.name}
                            className={cn(
                              "text-center",
                              product.highlighted && "bg-brand/5"
                            )}
                          >
                            <div className="flex justify-center">
                              <CellValue value={product.values[rowIdx]} />
                            </div>
                          </TableCell>
                        ))}
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
                <ScrollBar orientation="horizontal" />
              </ScrollArea>
            </div>

            {/* Desktop: full table without scroll wrapper */}
            <div className="hidden md:block">
              <Table>
                <TableHeader>
                  <TableRow className="border-border-subtle hover:bg-transparent">
                    <TableHead className="w-[280px] pl-6 text-sm uppercase tracking-wider font-semibold text-text-muted py-5">
                      Feature
                    </TableHead>
                    {COMPARISON_PRODUCTS.map((product) => (
                      <TableHead
                        key={product.name}
                        className={cn(
                          "text-center min-w-[180px] relative py-5",
                          product.highlighted
                            ? "text-brand font-semibold text-base"
                            : "text-sm uppercase tracking-wider font-semibold text-text-muted"
                        )}
                      >
                        <div className="flex items-center justify-center gap-1.5">
                          {product.highlighted && (
                            <Zap className="h-4 w-4" />
                          )}
                          {product.name}
                        </div>
                        {product.highlighted && (
                          <ShineBorder
                            shineColor={["#FF9900", "#FFB84D"]}
                            borderWidth={1}
                          />
                        )}
                      </TableHead>
                    ))}
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {COMPARISON_FEATURES.map((feature, rowIdx) => (
                    <TableRow
                      key={feature}
                      className="border-border-subtle hover:bg-surface-light/30"
                    >
                      <TableCell className="pl-6 text-base font-medium text-text-secondary py-5">
                        {feature}
                      </TableCell>
                      {COMPARISON_PRODUCTS.map((product) => (
                        <TableCell
                          key={product.name}
                          className={cn(
                            "text-center py-5",
                            product.highlighted && "bg-brand/5"
                          )}
                        >
                          <div className="flex justify-center">
                            <CellValue value={product.values[rowIdx]} />
                          </div>
                        </TableCell>
                      ))}
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          </GlassmorphicCard>
        </div>
      </BlurFade>
    </SectionWrapper>
  );
}
