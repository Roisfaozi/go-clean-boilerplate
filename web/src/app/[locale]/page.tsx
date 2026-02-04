import FAQ from "~/components/landing/faq";
import Features from "~/components/landing/features";
import Hero from "~/components/landing/hero";
import OpenSource from "~/components/landing/open-source";
import Pricing from "~/components/landing/pricing";
import Testimonials from "~/components/landing/testimonials";

export default async function Home() {
  return (
    <>
      <Hero />
      <Features />
      <Testimonials />
      <Pricing />
      <FAQ />
      <OpenSource />
    </>
  );
}
