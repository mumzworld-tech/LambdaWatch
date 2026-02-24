import { Badge } from "@/components/ui/badge";
import { SectionDivider, GitHubStarButton } from "@/components/common";
import { FOOTER_LINKS } from "@/lib/constants";

export function Footer() {
  return (
    <footer className="relative">
      <SectionDivider />
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 py-16">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-12">
          {/* Branding */}
          <div className="md:col-span-1">
            <span className="font-bold text-2xl text-brand">
              LambdaWatch
            </span>
            <p className="mt-3 text-sm text-text-muted max-w-xs">
              Ship Lambda logs to Grafana Loki in real-time. Zero code changes.
              Zero vendor lock-in.
            </p>
          </div>

          {/* Links */}
          <div className="grid grid-cols-2 gap-8 md:col-span-2">
            <div>
              <h3 className="text-sm font-semibold text-text-primary mb-4">
                Resources
              </h3>
              <ul className="space-y-3">
                {FOOTER_LINKS.resources.map((link) => (
                  <li key={link.label}>
                    <a
                      href={link.href}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-sm text-text-muted hover:text-text-primary transition-colors"
                    >
                      {link.label}
                    </a>
                  </li>
                ))}
              </ul>
            </div>
            <div>
              <h3 className="text-sm font-semibold text-text-primary mb-4">
                Community
              </h3>
              <ul className="space-y-3">
                {FOOTER_LINKS.community.map((link) => (
                  <li key={link.label}>
                    <a
                      href={link.href}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-sm text-text-muted hover:text-text-primary transition-colors"
                    >
                      {link.label}
                    </a>
                  </li>
                ))}
              </ul>
            </div>
          </div>
        </div>

        {/* Bottom bar */}
        <div className="mt-16 pt-8 border-t border-border-subtle flex flex-col sm:flex-row items-center justify-between gap-4">
          <div className="flex items-center gap-3">
            <span className="text-sm text-text-muted">
              Built with Go. Pure standard library.
            </span>
            <Badge
              variant="secondary"
              className="bg-surface-lighter text-text-muted border-border-subtle"
            >
              MIT
            </Badge>
          </div>
          <GitHubStarButton />
        </div>
      </div>
    </footer>
  );
}
