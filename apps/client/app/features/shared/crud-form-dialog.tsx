import { useState, useEffect } from "react";
import { NexusButton } from "@casbin/ui";
import { NexusInput } from "@casbin/ui";
import { NexusTextarea } from "@casbin/ui";
import { FormGroup } from "@/components/patterns/form-group";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
  DialogDescription,
} from "@casbin/ui";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@casbin/ui";
import { Switch } from "@casbin/ui";
import { Loader2 } from "lucide-react";
import type { z } from "zod";

/* ── Types ── */
export interface FieldDef {
  name: string;
  label: string;
  type:
    | "text"
    | "email"
    | "password"
    | "textarea"
    | "select"
    | "switch"
    | "number";
  placeholder?: string;
  required?: boolean;
  description?: string;
  options?: { label: string; value: string }[];
  disabled?: boolean;
  hidden?: boolean;
}

interface CrudFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: string;
  description?: string;
  fields: FieldDef[];
  schema: z.ZodObject<any>;
  initialValues?: Record<string, any>;
  onSubmit: (values: Record<string, any>) => Promise<void> | void;
  submitLabel?: string;
  loading?: boolean;
}

export function CrudFormDialog({
  open,
  onOpenChange,
  title,
  description,
  fields,
  schema,
  initialValues,
  onSubmit,
  submitLabel = "Save",
  loading,
}: CrudFormDialogProps) {
  const [values, setValues] = useState<Record<string, any>>({});
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (open) {
      const defaults: Record<string, any> = {};
      fields.forEach((f) => {
        defaults[f.name] =
          initialValues?.[f.name] ?? (f.type === "switch" ? false : "");
      });
      setValues(defaults);
      setErrors({});
    }
  }, [open, initialValues, fields]);

  const setValue = (name: string, value: any) => {
    setValues((prev) => ({ ...prev, [name]: value }));
    setErrors((prev) => {
      const n = { ...prev };
      delete n[name];
      return n;
    });
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const result = schema.safeParse(values);
    if (!result.success) {
      const fieldErrors: Record<string, string> = {};
      result.error.errors.forEach((err) => {
        const path = err.path[0] as string;
        if (!fieldErrors[path]) fieldErrors[path] = err.message;
      });
      setErrors(fieldErrors);
      return;
    }
    setSubmitting(true);
    try {
      await onSubmit(result.data);
      onOpenChange(false);
    } catch {
      // error handled by caller
    } finally {
      setSubmitting(false);
    }
  };

  const isLoading = submitting || loading;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          {description && <DialogDescription>{description}</DialogDescription>}
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4 py-2">
          {fields
            .filter((f) => !f.hidden)
            .map((field) => (
              <FormGroup
                key={field.name}
                label={field.label}
                required={field.required}
                error={errors[field.name]}
                description={field.description}
              >
                {field.type === "textarea" ? (
                  <NexusTextarea
                    id={field.name}
                    name={field.name}
                    value={values[field.name] || ""}
                    onChange={(e) => setValue(field.name, e.target.value)}
                    placeholder={field.placeholder}
                    disabled={isLoading || field.disabled}
                  />
                ) : field.type === "select" ? (
                  <Select
                    value={values[field.name] || ""}
                    onValueChange={(val) => setValue(field.name, val)}
                    disabled={isLoading || field.disabled}
                  >
                    <SelectTrigger>
                      <SelectValue
                        placeholder={
                          field.placeholder ||
                          `Select ${field.label.toLowerCase()}`
                        }
                      />
                    </SelectTrigger>
                    <SelectContent>
                      {field.options?.map((opt) => (
                        <SelectItem key={opt.value} value={opt.value}>
                          {opt.label}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                ) : field.type === "switch" ? (
                  <Switch
                    id={field.name}
                    checked={!!values[field.name]}
                    onCheckedChange={(val) => setValue(field.name, val)}
                    disabled={isLoading || field.disabled}
                  />
                ) : (
                  <NexusInput
                    id={field.name}
                    name={field.name}
                    type={field.type === "number" ? "number" : field.type}
                    value={values[field.name] || ""}
                    onChange={(e) =>
                      setValue(
                        field.name,
                        field.type === "number"
                          ? Number(e.target.value)
                          : e.target.value,
                      )
                    }
                    placeholder={field.placeholder}
                    disabled={isLoading || field.disabled}
                  />
                )}
              </FormGroup>
            ))}
          <DialogFooter className="pt-2">
            <NexusButton
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={isLoading}
            >
              Cancel
            </NexusButton>
            <NexusButton type="submit" disabled={isLoading}>
              {isLoading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              {submitLabel}
            </NexusButton>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
