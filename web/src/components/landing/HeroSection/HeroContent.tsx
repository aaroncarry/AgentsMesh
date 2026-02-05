"use client";

import Link from "next/link";
import { Button } from "@/components/ui/button";

interface HeroContentProps {
  t: (key: string) => string;
}

/**
 * HeroContent - Renders the hero section text content (left side)
 */
export function HeroContent({ t }: HeroContentProps) {
  return (
    <div className="text-center lg:text-left">
      {/* Badge */}
      <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-primary/10 border border-primary/20 text-primary text-sm mb-6">
        <span className="relative flex h-2 w-2">
          <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-primary opacity-75"></span>
          <span className="relative inline-flex rounded-full h-2 w-2 bg-primary"></span>
        </span>
        {t("landing.hero.badge")}
      </div>

      {/* Headline */}
      <h1 className="text-4xl sm:text-5xl lg:text-6xl font-extrabold leading-tight mb-4">
        <span className="text-foreground">{t("landing.hero.slogan1")}</span>
        <br />
        <span className="text-primary">{t("landing.hero.slogan2")}</span>
      </h1>

      {/* Tagline */}
      <p className="text-lg sm:text-xl text-muted-foreground/80 mb-6 font-medium italic">
        {t("landing.hero.tagline")}
      </p>

      {/* Description */}
      <p className="text-lg sm:text-xl text-muted-foreground mb-8 max-w-xl mx-auto lg:mx-0">
        {t("landing.hero.description")}
      </p>

      {/* CTA Buttons */}
      <div className="flex flex-col sm:flex-row gap-4 justify-center lg:justify-start">
        <Link href="/register">
          <Button size="lg" className="w-full sm:w-auto bg-primary text-primary-foreground hover:bg-primary/90 text-base px-8">
            {t("landing.hero.getStartedFree")}
          </Button>
        </Link>
        <Link href="/docs">
          <Button size="lg" variant="outline" className="w-full sm:w-auto text-base px-8">
            {t("landing.hero.viewDocs")}
          </Button>
        </Link>
      </div>

      {/* Trust badges */}
      <div className="mt-10 pt-8 border-t border-border/50">
        <p className="text-sm text-muted-foreground mb-4">{t("landing.hero.trustedBy")}</p>
        <div className="flex items-center justify-center lg:justify-start gap-6 opacity-50">
          <div className="text-sm font-medium">{t("landing.hero.teams")}</div>
          <div className="w-px h-4 bg-border" />
          <div className="text-sm font-medium">{t("landing.hero.pods")}</div>
          <div className="w-px h-4 bg-border" />
          <div className="text-sm font-medium">{t("landing.hero.openSource")}</div>
        </div>
      </div>
    </div>
  );
}

export default HeroContent;
