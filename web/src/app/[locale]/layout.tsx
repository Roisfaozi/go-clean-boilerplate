import { type Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import Script from "next/script";
import { AnnouncementBanner } from "~/components/layout/announcement-banner";
import Footer from "~/components/layout/footer";
import Header from "~/components/layout/header";
import ThemeProvider from "~/components/shared/providers/theme-provider";
import { DensityProvider } from "~/components/shared/providers/density-provider";
import { WebSocketProvider } from "~/components/shared/providers/websocket-provider";
import { AuthProvider } from "~/components/shared/providers/auth-provider";
import { Toaster } from "~/components/ui/toaster";
import { siteConfig, siteUrl } from "~/config/site";
import { cn } from "~/lib/utils";
import { I18nProviderClient } from "~/locales/client";
import "../globals.css";

const fontSans = Geist({
  subsets: ["latin"],
  variable: "--font-sans",
});

const fontMono = Geist_Mono({
  subsets: ["latin"],
  variable: "--font-mono",
});

type Props = {
  params: Promise<{ locale: string }>;
  searchParams: { [key: string]: string | string[] | undefined };
};

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const p = await params;
  const locale = p.locale;
  const site = siteConfig(locale);

  const siteOgImage = `${siteUrl}/api/og?locale=${locale}`;

  return {
    title: {
      default: site.name,
      template: `%s - ${site.name}`,
    },
    description: site.description,
    keywords: [
      "Next.js",
      "Shadcn/ui",
      "LuciaAuth",
      "Prisma",
      "Vercel",
      "Tailwind",
      "Radix UI",
      "Stripe",
      "Internationalization",
      "Postgres",
    ],
    authors: [
      {
        name: "moinulmoin",
        url: "https://moinulmoin.com",
      },
    ],
    creator: "Moinul Moin",
    openGraph: {
      type: "website",
      locale: locale,
      url: site.url,
      title: site.name,
      description: site.description,
      siteName: site.name,
      images: [
        {
          url: siteOgImage,
          width: 1200,
          height: 630,
          alt: site.name,
        },
      ],
    },
    twitter: {
      card: "summary_large_image",
      title: site.name,
      description: site.description,
      images: [siteOgImage],
      creator: "@immoinulmoin",
    },
    icons: {
      icon: "/favicon.ico",
      shortcut: "/favicon-16x16.png",
      apple: "/apple-touch-icon.png",
    },
    manifest: `${siteUrl}/manifest.json`,
    metadataBase: new URL(site.url),
    alternates: {
      canonical: "/",
      languages: {
        en: "/en",
        fr: "/fr",
      },
    },
    appleWebApp: {
      capable: true,
      statusBarStyle: "default",
      title: site.name,
    },
  };
}

export const viewport = {
  width: 1,
  themeColor: [
    { media: "(prefers-color-scheme: light)", color: "white" },
    { media: "(prefers-color-scheme: dark)", color: "black" },
  ],
};

export default async function RootLayout({
  children,
  loginDialog,
  params,
}: {
  children: React.ReactNode;
  loginDialog: React.ReactNode;
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;
  return (
    <html lang={locale} suppressHydrationWarning>
      <body
        className={cn(
          "font-sans antialiased",
          fontSans.variable,
          fontMono.variable
        )}
      >
        <ThemeProvider attribute="class" defaultTheme="system" enableSystem>
          <DensityProvider>
            <AuthProvider>
              <WebSocketProvider>
                <AnnouncementBanner />
                <Header />
                <main>
                  {children}
                  {loginDialog}
                </main>
                <I18nProviderClient locale={locale}>
                  <Footer />
                </I18nProviderClient>
                <Toaster />
              </WebSocketProvider>
            </AuthProvider>
          </DensityProvider>
        </ThemeProvider>
      </body>
      {/* remove this line when you are working on your own project */}
      {process.env.NODE_ENV === "production" && (
        <Script
          src="https://umami.moinulmoin.com/script.js"
          data-website-id="bc66d96a-fc75-4ecd-b0ef-fdd25de8113c"
        />
      )}
    </html>
  );
}
