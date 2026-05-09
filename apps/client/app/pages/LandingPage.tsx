import { LandingNavbar } from "@/components/landing/landing-navbar";
import { LandingHero } from "@/components/landing/landing-hero";
import { LandingPartners } from "@/components/landing/landing-partners";
import { LandingFeatures } from "@/components/landing/landing-features";
import { LandingFeatureSplit } from "@/components/landing/landing-feature-split";
import { LandingCtaBanner } from "@/components/landing/landing-cta-banner";
import { LandingTestimonials } from "@/components/landing/landing-testimonials";
import { LandingPricing } from "@/components/landing/landing-pricing";
import { LandingNewsletter } from "@/components/landing/landing-newsletter";
import { LandingFooter } from "@/components/landing/landing-footer";

export default function LandingPage() {
  return (
    <div className="bg-background text-foreground min-h-screen">
      <LandingNavbar />
      <main>
        <LandingHero />
        <LandingPartners />
        <LandingFeatures />
        <LandingFeatureSplit />
        <LandingCtaBanner
          variant="primary"
          eyebrow="Get started"
          title="Launch your internal platform faster with a system that already scales"
          description="Dari auth sampai access control, dari dashboard sampai marketing surface — semua siap dibangun di atas fondasi yang konsisten."
          primaryCta={{ label: "Start Building", to: "/register" }}
          secondaryCta={{ label: "Talk to Sales", to: "/login" }}
        />
        <LandingTestimonials />
        <LandingPricing />
        <LandingCtaBanner
          variant="surface"
          title="Ready to unify your product operations?"
          description="Mulai dengan fondasi yang rapi, sistem yang konsisten, dan arsitektur yang siap berkembang."
          primaryCta={{ label: "Start Free Trial", to: "/register" }}
          secondaryCta={{ label: "Request Demo", to: "/login" }}
        />
        <LandingNewsletter />
      </main>
      <LandingFooter />
    </div>
  );
}
