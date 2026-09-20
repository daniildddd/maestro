import { Button } from "./ui";

export function LiveToggle({
  active,
  onToggle,
}: {
  active: boolean;
  onToggle: () => void;
}) {
  return (
    <div className="live-indicator" title={active ? "Auto-refresh every few seconds. Click to pause background polling." : "Background polling is paused. Click to resume."}>
      <span className={"live-dot" + (active ? " on" : "")} aria-hidden="true" />
      <span>{active ? "Live" : "Polling paused"}</span>
      <Button variant="ghost" className="btn-sm" onClick={onToggle} type="button">
        {active ? "Pause polling" : "Resume polling"}
      </Button>
    </div>
  );
}