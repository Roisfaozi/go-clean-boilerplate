"use client";

import Link from "next/link";
import LogoutButton from "~/components/shared/logout-button";
import { Button, buttonVariants } from "~/components/ui/button";
import { Sheet, SheetContent, SheetTrigger } from "~/components/ui/sheet";
import { cn } from "~/lib/utils";
import Icons from "../shared/icons";
import { useState } from "react";
import { MenuIcon } from "lucide-react";
import DensityToggle from "../shared/density-toggle";

export default function Navbar({
  session,
  headerText,
}: {
  session: any;
  headerText: {
    changelog: string;
    about: string;
    login: string;
    dashboard: string;
    [key: string]: string;
  };
}) {
  const [isOpen, setIsOpen] = useState(false);

  return (
    <nav className="flex items-center justify-between py-4">
      <Link href="/" className="flex items-center gap-2">
        <Icons.logo className="h-8 w-8" />
        <span className="text-lg font-bold tracking-tight">NexusOS</span>
      </Link>

      {/* Desktop Nav */}
      <div className="hidden items-center gap-6 md:flex">
        <Link
          href="/changelog"
          className="text-muted-foreground hover:text-primary text-sm font-medium transition-colors"
        >
          {headerText.changelog}
        </Link>
        <Link
          href="/about"
          className="text-muted-foreground hover:text-primary text-sm font-medium transition-colors"
        >
          {headerText.about}
        </Link>

        <div className="h-4 w-px bg-slate-200 dark:bg-slate-800" />
        <DensityToggle />

        {session ? (
          <div className="flex items-center gap-4">
            <Link
              href="/dashboard"
              className={cn(buttonVariants({ variant: "ghost", size: "sm" }))}
            >
              {headerText.dashboard}
            </Link>
            <LogoutButton />
          </div>
        ) : (
          <Link href="/login" className={cn(buttonVariants({ size: "sm" }))}>
            {headerText.login}
          </Link>
        )}
      </div>

      {/* Mobile Nav */}
      <Sheet open={isOpen} onOpenChange={setIsOpen}>
        <SheetTrigger asChild className="md:hidden">
          <Button variant="ghost" size="icon">
            <MenuIcon className="h-5 w-5" />
            <span className="sr-only">Toggle menu</span>
          </Button>
        </SheetTrigger>
        <SheetContent side="right">
          <div className="flex flex-col gap-4 py-4">
            <Link
              href="/changelog"
              onClick={() => setIsOpen(false)}
              className="text-sm font-medium"
            >
              {headerText.changelog}
            </Link>
            <Link
              href="/about"
              onClick={() => setIsOpen(false)}
              className="text-sm font-medium"
            >
              {headerText.about}
            </Link>
            
            <div className="flex items-center justify-between border-t border-slate-200 pt-4 dark:border-slate-800">
              <span className="text-sm font-medium">Density Mode</span>
              <DensityToggle />
            </div>

            {session ? (
              <>
                <Link
                  href="/dashboard"
                  onClick={() => setIsOpen(false)}
                  className="text-sm font-medium"
                >
                  {headerText.dashboard}
                </Link>
                <LogoutButton />
              </>
            ) : (
              <Link
                href="/login"
                onClick={() => setIsOpen(false)}
                className={cn(buttonVariants({ className: "w-full" }))}
              >
                {headerText.login}
              </Link>
            )}
          </div>
        </SheetContent>
      </Sheet>
    </nav>
  );
}
