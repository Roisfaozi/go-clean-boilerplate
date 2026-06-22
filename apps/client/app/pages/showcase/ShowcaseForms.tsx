import { useState } from "react";
import { PageHeader } from "@/components/layout/page-header";
import { FormGroup } from "@/components/patterns/form-group";
import { NexusInput } from "@casbin/ui";
import { NexusTextarea } from "@casbin/ui";
import { FormSection, FieldGroup } from "@/components/forms/form-section";
import { DatePicker } from "@/components/forms/date-picker";
import { TimePicker } from "@/components/forms/time-picker";
import { FileUpload } from "@/components/forms/file-upload";
import { MultiSelect, TagInput } from "@/components/forms/multi-select";
import { StepperForm } from "@/components/forms/stepper-form";
import { NexusCard, NexusCardContent, NexusCardFooter } from "@casbin/ui";
import { NexusButton } from "@casbin/ui";

export default function ShowcaseForms() {
  const [date, setDate] = useState<Date>();
  const [time, setTime] = useState("09:00");
  const [multiVal, setMultiVal] = useState<string[]>(["react"]);
  const [tags, setTags] = useState(["typescript", "react"]);
  const [step, setStep] = useState(1);

  return (
    <div className="max-w-5xl space-y-10">
      <PageHeader
        title="Form Controls"
        description="Inputs, pickers, selects, tags, file upload, and stepper."
      />

      <FormSection title="Basic Inputs" description="Standard form elements">
        <FieldGroup layout="horizontal">
          <FormGroup label="Email" required>
            <NexusInput type="email" placeholder="you@company.com" />
          </FormGroup>
          <FormGroup label="Password" required>
            <NexusInput type="password" placeholder="••••••••" />
          </FormGroup>
        </FieldGroup>
        <FieldGroup layout="horizontal">
          <FormGroup label="Disabled">
            <NexusInput disabled placeholder="Cannot edit" />
          </FormGroup>
          <FormGroup label="With Error" error="This field is required">
            <NexusInput
              placeholder="Oops"
              className="border-danger focus-visible:ring-danger"
            />
          </FormGroup>
        </FieldGroup>
      </FormSection>

      <FormSection title="Textarea" description="Multi-line text input">
        <div className="max-w-lg">
          <FormGroup label="Description">
            <NexusTextarea placeholder="Enter details…" rows={3} />
          </FormGroup>
        </div>
      </FormSection>

      <FormSection title="Date & Time Pickers">
        <FieldGroup layout="horizontal">
          <FormGroup label="Date">
            <DatePicker value={date} onChange={setDate} />
          </FormGroup>
          <FormGroup label="Time">
            <TimePicker value={time} onChange={setTime} />
          </FormGroup>
        </FieldGroup>
      </FormSection>

      <FormSection title="MultiSelect & Tags">
        <FieldGroup layout="horizontal">
          <FormGroup label="Technologies">
            <MultiSelect
              options={[
                { value: "react", label: "React" },
                { value: "vue", label: "Vue" },
                { value: "svelte", label: "Svelte" },
                { value: "angular", label: "Angular" },
                { value: "solid", label: "Solid" },
              ]}
              value={multiVal}
              onChange={setMultiVal}
            />
          </FormGroup>
          <FormGroup label="Tags">
            <TagInput value={tags} onChange={setTags} />
          </FormGroup>
        </FieldGroup>
      </FormSection>

      <FormSection title="File Upload">
        <FileUpload accept="image/*,.pdf" multiple />
      </FormSection>

      <FormSection title="Stepper Form">
        <StepperForm
          steps={[
            { label: "Account", description: "Basic info" },
            { label: "Profile", description: "Details" },
            { label: "Review", description: "Confirm" },
          ]}
          currentStep={step}
        >
          <NexusCard>
            <NexusCardContent className="text-muted-foreground py-8 text-center">
              Step {step + 1} content goes here
            </NexusCardContent>
            <NexusCardFooter className="justify-end gap-2">
              <NexusButton
                variant="outline"
                onClick={() => setStep(Math.max(0, step - 1))}
                disabled={step === 0}
              >
                Back
              </NexusButton>
              <NexusButton
                onClick={() => setStep(Math.min(2, step + 1))}
                disabled={step === 2}
              >
                Next
              </NexusButton>
            </NexusCardFooter>
          </NexusCard>
        </StepperForm>
      </FormSection>
    </div>
  );
}
