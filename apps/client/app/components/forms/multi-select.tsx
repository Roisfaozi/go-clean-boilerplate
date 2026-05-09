import { useState, useRef, useCallback, KeyboardEvent } from "react";
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

export function MultiSelect({ options, value = [], onChange, placeholder = "Select...", className }: MultiSelectProps) {
  const [open, setOpen] = useState(false);
  const [search, setSearch] = useState("");
  const ref = useRef<HTMLDivElement>(null);

  const filtered = options.filter(
    (o) => o.label.toLowerCase().includes(search.toLowerCase()) && !value.includes(o.value)
  );
  const selected = options.filter((o) => value.includes(o.value));

  const toggle = (val: string) => {
    const next = value.includes(val) ? value.filter((v) => v !== val) : [...value, val];
    onChange?.(next);
  };

  return (
    <div ref={ref} className={cn("relative", className)}>
      <div
        onClick={() => setOpen(!open)}
        className="flex flex-wrap items-center gap-1.5 min-h-input rounded-md border border-border bg-background px-3 py-2 cursor-pointer"
      >
        {selected.map((s) => (
          <span
            key={s.value}
            className="inline-flex items-center gap-1 bg-primary/10 text-primary text-caption rounded-full px-2 py-0.5"
          >
            {s.label}
            <X
              className="h-3 w-3 cursor-pointer hover:text-danger"
              onClick={(e) => { e.stopPropagation(); toggle(s.value); }}
            />
          </span>
        ))}
        {selected.length === 0 && <span className="text-muted-foreground text-body">{placeholder}</span>}
        <ChevronDown className="ml-auto h-4 w-4 text-muted-foreground shrink-0" />
      </div>
      {open && (
        <div className="absolute z-50 top-full left-0 right-0 mt-1 bg-popover border border-border rounded-md shadow-md animate-fade-in max-h-60 overflow-auto">
          <div className="p-2">
            <input
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Search..."
              className="w-full rounded-md border border-border bg-background px-3 py-1.5 text-body focus:outline-none focus:ring-1 focus:ring-ring"
              autoFocus
            />
          </div>
          {filtered.length === 0 && (
            <p className="px-3 py-2 text-caption text-muted-foreground">No options</p>
          )}
          {filtered.map((o) => (
            <button
              key={o.value}
              onClick={() => toggle(o.value)}
              className="flex items-center gap-2 w-full px-3 py-2 text-body hover:bg-surface-hover transition-colors text-left"
            >
              {o.label}
            </button>
          ))}
          {value.length > 0 && (
            <>
              <div className="border-t border-border my-1" />
              <p className="px-3 py-1 text-caption text-muted-foreground">Selected</p>
              {selected.map((o) => (
                <button
                  key={o.value}
                  onClick={() => toggle(o.value)}
                  className="flex items-center gap-2 w-full px-3 py-2 text-body hover:bg-surface-hover transition-colors text-left text-primary"
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

export function TagInput({ value = [], onChange, placeholder = "Add tag…", className }: TagInputProps) {
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
        "flex flex-wrap items-center gap-1.5 min-h-input rounded-md border border-border bg-background px-3 py-2",
        className
      )}
    >
      {value.map((tag) => (
        <span
          key={tag}
          className="inline-flex items-center gap-1 bg-primary/10 text-primary text-caption rounded-full px-2 py-0.5"
        >
          {tag}
          <X
            className="h-3 w-3 cursor-pointer hover:text-danger"
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
        className="flex-1 min-w-[80px] bg-transparent text-body outline-none placeholder:text-muted-foreground"
      />
    </div>
  );
}
