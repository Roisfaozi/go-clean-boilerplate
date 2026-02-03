"use client";

import Link from "next/link";
import LogoutButton from "~/components/shared/logout-button";
import { Button, buttonVariants } from "~/components/ui/button";
import { Sheet, SheetContent, SheetTrigger } from "~/components/ui/sheet";
import { cn } from "~/lib/utils";
import Icons from "../shared/icons";
import { useState } from "react";
import { MenuIcon } from "lucide-react";

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
          className="text-sm font-medium text-muted-foreground transition-colors hover:text-primary"
        >
          {headerText.changelog}
        </Link>
        <Link
          href="/about"
          className="text-sm font-medium text-muted-foreground transition-colors hover:text-primary"
        >
          {headerText.about}
        </Link>
        
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
          <Link
            href="/login"
            className={cn(buttonVariants({ size: "sm" }))}
          >
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
