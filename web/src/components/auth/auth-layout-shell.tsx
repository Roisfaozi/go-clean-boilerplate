"use client";

import Link from "next/link";
import Icons from "~/components/shared/icons";
import { ReactNode } from "react";

interface AuthLayoutShellProps {
  children: ReactNode;
  title: string;
  description: string;
  footer?: ReactNode;
  brandingTitle: string;
  brandingDescription: string;
  testimonial?: {
    quote: string;
    author: string;
    role: string;
  };
  features?: string[];
}

export function AuthLayoutShell({
  children,
  title,
  description,
  footer,
  brandingTitle,
  brandingDescription,
  testimonial,
  features,
}: AuthLayoutShellProps) {
  return (
    <div className="flex min-h-screen w-full flex-col lg:flex-row">
      {/* Left Panel: Functional Zone */}
      <div className="flex flex-1 flex-col justify-center px-6 py-12 md:px-12 lg:px-24 xl:px-32">
        <div className="mx-auto w-full max-w-sm lg:mx-0">
          <Link
            href="/"
            className="mb-10 flex items-center gap-2 transition-opacity hover:opacity-80"
          >
            <Icons.logo className="text-primary h-10 w-10" />
            <span className="text-2xl font-bold tracking-tighter">NexusOS</span>
          </Link>

          <div className="mb-8">
            <h1 className="text-3xl font-bold tracking-tight">{title}</h1>
            <p className="text-muted-foreground mt-2 text-sm">{description}</p>
          </div>

          {children}

          {footer && (
            <div className="mt-8 text-center text-sm lg:text-left">
              {footer}
            </div>
          )}
        </div>
      </div>

      {/* Right Panel: Visual/Branding Zone */}
      <div className="bg-primary relative hidden w-full flex-1 items-center justify-center overflow-hidden lg:flex">
        {/* Background Decorative Elements */}
        <div className="absolute inset-0 bg-linear-to-br from-indigo-600 to-violet-700 opacity-90" />
        <div className="absolute inset-0 bg-[url('https://www.transparenttextures.com/patterns/cubes.png')] opacity-20" />

        <div className="relative z-10 p-12 text-white">
          <div className="max-w-md">
            <div className="mb-8 flex h-12 w-12 items-center justify-center rounded-xl bg-white/20 backdrop-blur-md">
              <Icons.logo className="h-8 w-8" />
            </div>
            <h2 className="mb-6 text-4xl leading-tight font-bold tracking-tight">
              {brandingTitle}
            </h2>
            <p className="mb-10 text-lg text-indigo-100">
              {brandingDescription}
            </p>

            {testimonial && (
              <div className="rounded-2xl bg-white/10 p-6 backdrop-blur-lg">
                <p className="text-indigo-50 italic">
                  &quot;{testimonial.quote}&quot;
                </p>
                <div className="mt-4 flex items-center gap-3">
                  <div className="h-10 w-10 rounded-full bg-indigo-300/50" />
                  <div>
                    <p className="text-sm font-semibold">
                      {testimonial.author}
                    </p>
                    <p className="text-xs text-indigo-200">
                      {testimonial.role}
                    </p>
                  </div>
                </div>
              </div>
            )}

            {features && (
              <ul className="space-y-4">
                {features.map((item) => (
                  <li key={item} className="flex items-center gap-3">
                    <div className="flex h-5 w-5 items-center justify-center rounded-full bg-indigo-400/30 text-indigo-200">
                      <svg
                        xmlns="http://www.w3.org/2000/svg"
                        viewBox="0 0 20 20"
                        fill="currentColor"
                        className="h-3 w-3"
                      >
                        <path
                          fillRule="evenodd"
                          d="M16.704 4.153a.75.75 0 01.143 1.052l-8 10.5a.75.75 0 01-1.127.075l-4.5-4.5a.75.75 0 011.06-1.06l3.894 3.893 7.48-9.817a.75.75 0 011.05-.143z"
                          clipRule="evenodd"
                        />
                      </svg>
                    </div>
                    <span className="text-sm font-medium text-indigo-50">
                      {item}
                    </span>
                  </li>
                ))}
              </ul>
            )}
          </div>
        </div>

        {/* Animated Orbs */}
        <div className="absolute -right-24 -bottom-24 h-96 w-96 animate-pulse rounded-full bg-violet-500/20 blur-3xl" />
        <div className="absolute -top-24 -left-24 h-96 w-96 animate-pulse rounded-full bg-indigo-400/20 blur-3xl" />
      </div>
    </div>
  );
}
