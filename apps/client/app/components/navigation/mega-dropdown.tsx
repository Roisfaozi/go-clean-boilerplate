import { cn } from "@casbin/ui";
import { useState, useRef, useEffect } from "react";

interface MegaDropdownSection {
  title: string;
  items: { label: string; description?: string; href?: string; icon?: React.ReactNode; onClick?: () => void }[];
}

interface MegaDropdownProps {
  trigger: React.ReactNode;
  sections: MegaDropdownSection[];
  footer?: React.ReactNode;
  className?: string;
}

export function MegaDropdown({ trigger, sections, footer, className }: MegaDropdownProps) {
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false);
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, []);

  return (
    <div ref={ref} className="relative">
      <div onClick={() => setOpen(!open)} className="cursor-pointer">
        {trigger}
      </div>
      {open && (
        <div
          className={cn(
            "absolute top-full left-0 mt-2 z-50 min-w-[480px] bg-popover border border-border rounded-lg shadow-lg p-card-pad animate-fade-in",
            className
          )}
        >
          <div className="grid grid-cols-2 gap-6">
            {sections.map((section) => (
              <div key={section.title}>
                <p className="text-caption font-semibold uppercase tracking-wider text-muted-foreground mb-2">
                  {section.title}
                </p>
                <div className="space-y-1">
                  {section.items.map((item) => (
                    <a
                      key={item.label}
                      href={item.href || "#"}
                      onClick={(e) => {
                        if (item.onClick) {
                          e.preventDefault();
                          item.onClick();
                        }
                        setOpen(false);
                      }}
                      className="flex items-start gap-3 rounded-md p-2 hover:bg-surface-hover transition-colors"
                    >
                      {item.icon && <span className="mt-0.5 shrink-0 text-muted-foreground">{item.icon}</span>}
                      <div>
                        <p className="text-body font-medium text-foreground">{item.label}</p>
                        {item.description && (
                          <p className="text-caption text-muted-foreground">{item.description}</p>
                        )}
                      </div>
                    </a>
                  ))}
                </div>
              </div>
            ))}
          </div>
          {footer && (
            <div className="mt-4 pt-4 border-t border-border">{footer}</div>
          )}
        </div>
      )}
    </div>
  );
}
