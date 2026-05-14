import Marquee from "~/components/magicui/marquee";
import Image from "next/image";

const tech = [
  {
    name: "Next.js 16",
    logo: "https://cdn.worldvectorlogo.com/logos/next-js.svg",
  },
  {
    name: "Golang",
    logo: "https://cdn.worldvectorlogo.com/logos/golang-1.svg",
  },
  {
    name: "Tailwind v4",
    logo: "https://cdn.worldvectorlogo.com/logos/tailwindcss.svg",
  },
  { name: "Casbin", logo: "https://casbin.org/img/casbin.svg" },
  { name: "MySQL", logo: "https://cdn.worldvectorlogo.com/logos/mysql-6.svg" },
  { name: "Redis", logo: "https://cdn.worldvectorlogo.com/logos/redis.svg" },
  { name: "Docker", logo: "https://cdn.worldvectorlogo.com/logos/docker.svg" },
  {
    name: "TypeScript",
    logo: "https://cdn.worldvectorlogo.com/logos/typescript.svg",
  },
];

export default function TechStack() {
  return (
    <section className="border-y border-slate-200 py-12 dark:border-slate-800">
      <div className="container px-4 md:px-6">
        <p className="mb-8 text-center text-sm font-medium tracking-widest text-slate-500 uppercase">
          Powered by industry-standard tech
        </p>
        <Marquee pauseOnHover className="[--duration:20s]">
          {tech.map((item) => (
            <div
              key={item.name}
              className="flex cursor-default items-center gap-2 px-8 opacity-50 grayscale transition-all hover:opacity-100 hover:grayscale-0"
            >
              <Image
                src={item.logo}
                alt={item.name}
                height={32}
                width={32}
                className="h-8 w-auto"
              />
              <span className="text-xl font-bold text-slate-700 dark:text-slate-300">
                {item.name}
              </span>
            </div>
          ))}
        </Marquee>
      </div>
    </section>
  );
}
