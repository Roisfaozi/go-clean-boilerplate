"use client";

import { useEffect, useRef, useState } from "react";
import { Input } from "~/components/ui/input";

interface SearchInputProps {
  defaultValue?: string;
  onSearch: (value: string) => void;
  placeholder?: string;
  className?: string;
  debounceMs?: number;
}

export function SearchInput({
  defaultValue = "",
  onSearch,
  placeholder = "Search...",
  className,
  debounceMs = 300,
}: SearchInputProps) {
  const [value, setValue] = useState(defaultValue);
  const onSearchRef = useRef(onSearch);

  useEffect(() => {
    onSearchRef.current = onSearch;
  }, [onSearch]);

  useEffect(() => {
    setValue(defaultValue);
  }, [defaultValue]);

  useEffect(() => {
    const handler = setTimeout(() => {
      onSearchRef.current(value);
    }, debounceMs);
    return () => clearTimeout(handler);
  }, [value, debounceMs]);

  return (
    <Input
      placeholder={placeholder}
      value={value}
      onChange={(e) => setValue(e.target.value)}
      className={className}
    />
  );
}
