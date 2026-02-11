import FAQ from "~/components/landing/faq";
import Features from "~/components/landing/features";
import Hero from "~/components/landing/hero";
import OpenSource from "~/components/landing/open-source";
import Pricing from "~/components/landing/pricing";
import TechStack from "~/components/landing/tech-stack";
import Testimonials from "~/components/landing/testimonials";
import Header from "~/components/layout/header";

export default async function Home() {
  return (
    <>
      <Header />

      <Hero />
      <TechStack />
      <Features />
      <Testimonials />
      <Pricing />
      <FAQ />
      <OpenSource />
    </>
  );
}
