import { Check } from "lucide-react";

export function CheckCard({
  checked,
  onChange,
  title,
  desc,
  disabled = false,
}: {
  checked: boolean;
  onChange: (v: boolean) => void;
  title: string;
  desc?: string;
  disabled?: boolean;
}) {
  return (
    <label
      className={"check-card" + (checked ? " checked" : "") + (disabled ? " disabled" : "")}
      title={disabled ? "Requires Include tasks" : undefined}
    >
      <input
        type="checkbox"
        checked={checked}
        disabled={disabled}
        onChange={(e) => onChange(e.target.checked)}
      />
      <span className="box" aria-hidden="true">
        <Check size={13} strokeWidth={3.5} />
      </span>
      <span className="check-text">
        <span className="check-title">{title}</span>
        {desc ? <span className="check-desc">{desc}</span> : null}
      </span>
    </label>
  );
}