"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { HeroContent } from "./HeroContent";
import { DemoVideoModal } from "./DemoVideoModal";

export function HeroSection() {
  const t = useTranslations();
  const [demoOpen, setDemoOpen] = useState(false);

  return (
    <section className="relative pt-32 pb-24 sm:pt-40 sm:pb-32 px-4 sm:px-6 lg:px-8 azure-mesh-bg overflow-hidden">
      <div className="absolute -top-20 -right-20 w-[500px] h-[500px] bg-[var(--azure-cyan)]/10 blur-[120px] rounded-full azure-orb pointer-events-none" />
      <div
        className="absolute bottom-10 -left-10 w-[400px] h-[400px] bg-[var(--azure-mint)]/10 blur-[100px] rounded-full azure-orb pointer-events-none"
        style={{ animationDelay: "1.5s" }}
      />

      <div
        className="absolute inset-0 opacity-[0.05] pointer-events-none"
        style={{
          backgroundImage:
            "linear-gradient(var(--azure-cyan) 1px, transparent 1px), linear-gradient(90deg, var(--azure-cyan) 1px, transparent 1px)",
          backgroundSize: "80px 80px",
          maskImage: "radial-gradient(ellipse at center, black 0%, transparent 70%)",
          WebkitMaskImage: "radial-gradient(ellipse at center, black 0%, transparent 70%)",
        }}
      />

      <div className="relative z-10 max-w-6xl mx-auto">
        <HeroContent t={t} onWatchDemo={() => setDemoOpen(true)} />
      </div>

      <DemoVideoModal
        open={demoOpen}
        onClose={() => setDemoOpen(false)}
        iframeTitle={t("landing.demoVideo.iframeTitle")}
      />
    </section>
  );
}

export default HeroSection;
