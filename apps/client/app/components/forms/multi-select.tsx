import { useState, useRef, type KeyboardEvent } from "react";
import { cn } from "@casbin/ui";
import { X, Check, ChevronDown } from "lucide-react";

/* ── MultiSelect ── */
export interface MultiSelectOption {
  value: string;
  label: string;
}

interface MultiSelectProps {
  options: MultiSelectOption[];
  value?: string[];
  onChange?: (values: string[]) => void;
  placeholder?: string;
  className?: string;
}

export function MultiSelect({
  options,
  value = [],
  onChange,
  placeholder = "Select...",
  className,
}: MultiSelectProps) {
  const [open, setOpen] = useState(false);
  const [search, setSearch] = useState("");
  const ref = useRef<HTMLDivElement>(null);

  const filtered = options.filter(
    (o) =>
      o.label.toLowerCase().includes(search.toLowerCase()) &&
      !value.includes(o.value),
  );
  const selected = options.filter((o) => value.includes(o.value));

  const toggle = (val: string) => {
    const next = value.includes(val)
      ? value.filter((v) => v !== val)
      : [...value, val];
    onChange?.(next);
  };

  return (
    <div ref={ref} className={cn("relative", className)}>
      <div
        onClick={() => setOpen(!open)}
        className="min-h-input border-border bg-background flex cursor-pointer flex-wrap items-center gap-1.5 rounded-md border px-3 py-2"
      >
        {selected.map((s) => (
          <span
            key={s.value}
            className="bg-primary/10 text-primary text-caption inline-flex items-center gap-1 rounded-full px-2 py-0.5"
          >
            {s.label}
            <X
              className="hover:text-danger h-3 w-3 cursor-pointer"
              onClick={(e) => {
                e.stopPropagation();
                toggle(s.value);
              }}
            />
          </span>
        ))}
        {selected.length === 0 && (
          <span className="text-muted-foreground text-body">{placeholder}</span>
        )}
        <ChevronDown className="text-muted-foreground ml-auto h-4 w-4 shrink-0" />
      </div>
      {open && (
        <div className="bg-popover border-border animate-fade-in absolute top-full right-0 left-0 z-50 mt-1 max-h-60 overflow-auto rounded-md border shadow-md">
          <div className="p-2">
            <input
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Search..."
              className="border-border bg-background text-body focus:ring-ring w-full rounded-md border px-3 py-1.5 focus:ring-1 focus:outline-none"
              autoFocus
            />
          </div>
          {filtered.length === 0 && (
            <p className="text-caption text-muted-foreground px-3 py-2">
              No options
            </p>
          )}
          {filtered.map((o) => (
            <button
              key={o.value}
              onClick={() => toggle(o.value)}
              className="text-body hover:bg-surface-hover flex w-full items-center gap-2 px-3 py-2 text-left transition-colors"
            >
              {o.label}
            </button>
          ))}
          {value.length > 0 && (
            <>
              <div className="border-border my-1 border-t" />
              <p className="text-caption text-muted-foreground px-3 py-1">
                Selected
              </p>
              {selected.map((o) => (
                <button
                  key={o.value}
                  onClick={() => toggle(o.value)}
                  className="text-body hover:bg-surface-hover text-primary flex w-full items-center gap-2 px-3 py-2 text-left transition-colors"
                >
                  <Check className="h-3.5 w-3.5" />
                  {o.label}
                </button>
              ))}
            </>
          )}
        </div>
      )}
    </div>
  );
}

/* ── TagInput ── */
interface TagInputProps {
  value?: string[];
  onChange?: (tags: string[]) => void;
  placeholder?: string;
  className?: string;
}

export function TagInput({
  value = [],
  onChange,
  placeholder = "Add tag…",
  className,
}: TagInputProps) {
  const [input, setInput] = useState("");

  const addTag = () => {
    const tag = input.trim();
    if (tag && !value.includes(tag)) {
      onChange?.([...value, tag]);
    }
    setInput("");
  };

  const handleKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter" || e.key === ",") {
      e.preventDefault();
      addTag();
    } else if (e.key === "Backspace" && !input && value.length) {
      onChange?.(value.slice(0, -1));
    }
  };

  return (
    <div
      className={cn(
        "min-h-input border-border bg-background flex flex-wrap items-center gap-1.5 rounded-md border px-3 py-2",
        className,
      )}
    >
      {value.map((tag) => (
        <span
          key={tag}
          className="bg-primary/10 text-primary text-caption inline-flex items-center gap-1 rounded-full px-2 py-0.5"
        >
          {tag}
          <X
            className="hover:text-danger h-3 w-3 cursor-pointer"
            onClick={() => onChange?.(value.filter((t) => t !== tag))}
          />
        </span>
      ))}
      <input
        value={input}
        onChange={(e) => setInput(e.target.value)}
        onKeyDown={handleKeyDown}
        onBlur={addTag}
        placeholder={value.length === 0 ? placeholder : ""}
        className="text-body placeholder:text-muted-foreground min-w-[80px] flex-1 bg-transparent outline-none"
      />
    </div>
  );
}
