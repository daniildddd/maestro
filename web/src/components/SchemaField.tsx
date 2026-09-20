import type { PluginSchemaField } from "../api/types";
import { Field } from "./ui";

export function SchemaFieldInput({
  field,
  value,
  onChange,
  error,
}: {
  field: PluginSchemaField;
  value: string;
  onChange: (v: string) => void;
  error: string | null;
}) {
  const label = field.label || field.name;
  const isEnum = (field.values?.length ?? 0) > 0;
  const type = field.type.toUpperCase();

  if (isEnum) {
    return (
      <Field label={label} hint={field.description} error={error} required={field.required}>
        <select data-field={field.name} value={value} onChange={(e) => onChange(e.target.value)}>
          <option value="">— select —</option>
          {field.values!.map((v) => (
            <option key={v} value={v}>
              {v}
            </option>
          ))}
        </select>
      </Field>
    );
  }

  if (type === "BOOLEAN") {
    return (
      <Field label={label} hint={field.description} error={error} required={field.required}>
        <select data-field={field.name} value={value} onChange={(e) => onChange(e.target.value)}>
          <option value="">— select —</option>
          <option value="true">true</option>
          <option value="false">false</option>
        </select>
      </Field>
    );
  }

  if (type === "PASSWORD") {
    return (
      <Field label={label} hint={field.description} error={error} required={field.required}>
        <input
          data-field={field.name}
          type="password"
          value={value}
          onChange={(e) => onChange(e.target.value)}
          placeholder={field.default ?? undefined}
          className="mono"
          autoComplete="new-password"
        />
      </Field>
    );
  }

  const inputType = type === "INT" || type === "LONG" || type === "SHORT" ? "number" : "text";

  return (
    <Field label={label} hint={field.description} error={error} required={field.required}>
      <input
        data-field={field.name}
        type={inputType}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={field.default ?? undefined}
        className="mono"
      />
    </Field>
  );
}

// value input for a custom property whose key matches a hidden schema field:
// renders the proper control (enum/boolean select, password, number) instead of plain text
export function CustomValueInput({
  field,
  value,
  onChange,
}: {
  field: PluginSchemaField;
  value: string;
  onChange: (v: string) => void;
}) {
  const isEnum = (field.values?.length ?? 0) > 0;
  const type = field.type.toUpperCase();

  if (isEnum) {
    return (
      <select data-field={field.name} value={value} onChange={(e) => onChange(e.target.value)}>
        <option value="">— select —</option>
        {field.values!.map((v) => (
          <option key={v} value={v}>
            {v}
          </option>
        ))}
      </select>
    );
  }

  if (type === "BOOLEAN") {
    return (
      <select data-field={field.name} value={value} onChange={(e) => onChange(e.target.value)}>
        <option value="">— select —</option>
        <option value="true">true</option>
        <option value="false">false</option>
      </select>
    );
  }

  if (type === "PASSWORD") {
    return (
      <input
        data-field={field.name}
        type="password"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="mono"
        autoComplete="new-password"
      />
    );
  }

  const inputType = type === "INT" || type === "LONG" || type === "SHORT" ? "number" : "text";

  return (
    <input
      data-field={field.name}
      type={inputType}
      value={value}
      onChange={(e) => onChange(e.target.value)}
      className="mono"
    />
  );
}