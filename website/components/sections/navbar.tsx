"use client";

import { useState, useCallback } from "react";
import { motion, useScroll, useTransform, AnimatePresence } from "motion/react";
import { Menu, X } from "lucide-react";
import { cn } from "@/lib/utils";
import { NAV_LINKS, GITHUB_URL } from "@/lib/constants";
import { GitHubStarButton } from "@/components/common";

export function Navbar() {
  const [mobileOpen, setMobileOpen] = useState(false);
  const { scrollY } = useScroll();
  const bgOpacity = useTransform(scrollY, [0, 100], [0, 0.7]);
  const borderOpacity = useTransform(scrollY, [0, 100], [0, 0.08]);
  const backdropBlur = useTransform(scrollY, [0, 100], [0, 12]);

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
        className="fixed top-0 left-0 right-0 z-50 h-16"
        style={{
          backgroundColor: useTransform(
            bgOpacity,
            (v) => `rgba(10, 10, 10, ${v})`
          ),
          borderBottom: useTransform(
            borderOpacity,
            (v) => `1px solid rgba(255, 153, 0, ${v})`
          ),
          backdropFilter: useTransform(
            backdropBlur,
            (v) => `blur(${v}px)`
          ),
        }}
      >
        <nav className="mx-auto flex h-full max-w-7xl items-center justify-between px-4 sm:px-6 lg:px-8">
          {/* Logo */}
          <a
            href="#"
            onClick={(e) => {
              e.preventDefault();
              window.scrollTo({ top: 0, behavior: "smooth" });
            }}
            className="font-bold text-xl text-brand transition-opacity hover:opacity-80"
          >
            LambdaWatch
          </a>

          {/* Center: Desktop nav links */}
          <div className="hidden items-center gap-1 md:flex">
            {NAV_LINKS.map((link) => (
              <a
                key={link.href}
                href={link.href}
                onClick={(e) => handleNavClick(e, link.href)}
                className="rounded-md px-3 py-2 text-sm font-medium text-text-secondary transition-colors hover:text-text-primary"
              >
                {link.label}
              </a>
            ))}
          </div>

          {/* Right: Desktop CTAs */}
          <div className="hidden items-center gap-3 md:flex">
            <GitHubStarButton className="py-2 text-xs" />
            <a
              href="#features"
              onClick={(e) => handleNavClick(e, "#features")}
              className="inline-flex items-center rounded-lg bg-brand px-4 py-2 text-sm font-semibold text-black transition-colors hover:bg-brand-light"
            >
              Get Started
            </a>
          </div>

          {/* Mobile: Hamburger */}
          <button
            type="button"
            className="inline-flex items-center justify-center rounded-md p-2 text-text-secondary transition-colors hover:text-text-primary md:hidden"
            onClick={() => setMobileOpen(true)}
            aria-label="Open navigation menu"
          >
            <Menu className="h-6 w-6" />
          </button>
        </nav>
      </motion.header>

      {/* Mobile overlay */}
      <AnimatePresence>
        {mobileOpen && (
          <motion.div
            className="fixed inset-0 z-[60] flex flex-col bg-black/95 backdrop-blur-lg md:hidden"
            initial={{ opacity: 0, x: "100%" }}
            animate={{ opacity: 1, x: 0 }}
            exit={{ opacity: 0, x: "100%" }}
            transition={{ type: "spring", damping: 25, stiffness: 200 }}
          >
            {/* Mobile header */}
            <div className="flex h-16 items-center justify-between px-4 sm:px-6">
              <span className="font-bold text-xl text-brand">
                LambdaWatch
              </span>
              <button
                type="button"
                className="inline-flex items-center justify-center rounded-md p-2 text-text-secondary transition-colors hover:text-text-primary"
                onClick={() => setMobileOpen(false)}
                aria-label="Close navigation menu"
              >
                <X className="h-6 w-6" />
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
                <GitHubStarButton />
                <a
                  href="#features"
                  onClick={(e) => handleNavClick(e, "#features")}
                  className="inline-flex items-center rounded-lg bg-brand px-6 py-3 text-base font-semibold text-black transition-colors hover:bg-brand-light"
                >
                  Get Started
                </a>
              </div>
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </>
  );
}
