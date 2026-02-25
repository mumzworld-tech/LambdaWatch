"use client";

import { useState, useCallback } from "react";
import { motion, useScroll, useTransform, AnimatePresence } from "motion/react";
import { Menu, X } from "lucide-react";
import Image from "next/image";
import { cn } from "@/lib/utils";
import { NAV_LINKS } from "@/lib/constants";
import { GitHubStarButton } from "@/components/common";

interface NavbarProps {
  stars?: number | null;
}

export function Navbar({ stars }: NavbarProps) {
  const [mobileOpen, setMobileOpen] = useState(false);
  const { scrollY } = useScroll();

  const borderOpacity = useTransform(scrollY, [0, 100], [0.08, 0.15]);
  const shadow = useTransform(
    scrollY,
    [0, 100],
    [
      "0 0 0 0 rgba(0,0,0,0)",
      "0 8px 32px rgba(0,0,0,0.4), 0 0 0 1px rgba(255,153,0,0.06)",
    ]
  );

  const handleNavClick = useCallback(
    (e: React.MouseEvent<HTMLAnchorElement>, href: string) => {
      e.preventDefault();
      setMobileOpen(false);
      const target = document.querySelector(href);
      if (target) {
        target.scrollIntoView({ behavior: "smooth" });
      }
    },
    []
  );

  return (
    <>
      <motion.header
        className={cn(
          "fixed top-3 sm:top-4 left-1/2 z-50 py-3 max-w-5xl w-[calc(100%-2rem)] -translate-x-1/2 rounded-2xl",
          "bg-clip-padding backdrop-filter backdrop-blur-xl"
        )}
        style={{
          border: useTransform(
            borderOpacity,
            (v) => `1px solid rgba(255, 153, 0, ${v})`
          ),
          boxShadow: shadow,
        }}
      >
        <nav className="mx-auto flex items-center justify-between px-4 sm:px-5">
          {/* Logo */}
          <a
            href="#"
            onClick={(e) => {
              e.preventDefault();
              window.scrollTo({ top: 0, behavior: "smooth" });
            }}
            className="flex items-center gap-2 font-bold text-lg text-brand transition-opacity hover:opacity-80"
          >
            <div className="flex items-center justify-center h-9 w-9 rounded-lg bg-surface-lighter border border-border-subtle">
              <Image
                src={`${process.env.NEXT_PUBLIC_BASE_PATH || ""}/logo.png`}
                alt="LambdaWatch logo"
                className="h-5 w-5"
                width={20}
                height={20}
              />
            </div>
            <span>LambdaWatch</span>
          </a>

          {/* Center: Desktop nav links */}
          <div className="hidden items-center gap-0.5 lg:flex">
            {NAV_LINKS.map((link) => (
              <a
                key={link.href}
                href={link.href}
                onClick={(e) => handleNavClick(e, link.href)}
                className="rounded-lg px-3 py-1.5 text-sm font-medium text-text-secondary transition-colors hover:text-text-primary hover:bg-white/5"
              >
                {link.label}
              </a>
            ))}
          </div>

          {/* Right: Desktop CTA */}
          <div className="hidden items-center md:flex md:ml-auto lg:ml-0">
            <GitHubStarButton stars={stars} className="rounded-full" />
          </div>

          {/* Mobile: Hamburger */}
          <button
            type="button"
            className="inline-flex ms-3 items-center justify-center rounded-lg p-2 text-text-secondary transition-colors hover:text-text-primary hover:bg-white/5 lg:hidden"
            onClick={() => setMobileOpen(true)}
            aria-label="Open navigation menu"
          >
            <Menu className="h-5 w-5" />
          </button>
        </nav>
      </motion.header>

      {/* Mobile overlay */}
      <AnimatePresence>
        {mobileOpen && (
          <motion.div
            className="fixed inset-0 z-[60] flex flex-col bg-black/95 backdrop-blur-xl lg:hidden"
            initial={{ opacity: 0, x: "100%" }}
            animate={{ opacity: 1, x: 0 }}
            exit={{ opacity: 0, x: "100%" }}
            transition={{ type: "spring", damping: 25, stiffness: 200 }}
          >
            {/* Mobile header */}
            <div className="flex min-h-14 pt-6 items-center justify-between px-4 sm:px-6">
              <a
                href="#"
                onClick={(e) => {
                  e.preventDefault();
                  setMobileOpen(false);
                  window.scrollTo({ top: 0, behavior: "smooth" });
                }}
                className="flex items-center gap-2 font-bold text-lg text-brand"
              >
                <div className="flex items-center justify-center h-9 w-9 rounded-lg bg-surface-lighter border border-border-subtle">
                  <Image
                    src={`${process.env.NEXT_PUBLIC_BASE_PATH || ""}/logo.png`}
                    alt="LambdaWatch logo"
                    className="h-5 w-5"
                    width={20}
                    height={20}
                  />
                </div>
                <span>LambdaWatch</span>
              </a>
              <button
                type="button"
                className="inline-flex items-center justify-center rounded-lg p-2 text-text-secondary transition-colors hover:text-text-primary"
                onClick={() => setMobileOpen(false)}
                aria-label="Close navigation menu"
              >
                <X className="h-5 w-5" />
              </button>
            </div>

            {/* Mobile nav links */}
            <div className="flex flex-1 flex-col items-center justify-center gap-6">
              {NAV_LINKS.map((link, i) => (
                <motion.a
                  key={link.href}
                  href={link.href}
                  onClick={(e) => handleNavClick(e, link.href)}
                  className="text-2xl font-medium text-text-secondary transition-colors hover:text-text-primary"
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: 0.1 + i * 0.05 }}
                >
                  {link.label}
                </motion.a>
              ))}

              <div className="mt-6 flex flex-col items-center gap-4">
                <GitHubStarButton stars={stars} />
              </div>
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </>
  );
}
