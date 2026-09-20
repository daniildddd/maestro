import { useEffect, useRef, useState } from "react";

// Periodically re-runs fn while the tab is visible. Pauses automatically when
// the document is hidden or enabled=false. Pass refreshKey to force a new
// schedule without changing the interval.
export function usePolling(fn: () => void, intervalMs: number, enabled = true) {
  const [polling, setPolling] = useState(enabled);
  const fnRef = useRef(fn);
  fnRef.current = fn;

  useEffect(() => {
    setPolling(enabled);
  }, [enabled]);

  useEffect(() => {
    if (!polling) return;
    const id = window.setInterval(() => {
      if (!document.hidden) fnRef.current();
    }, intervalMs);
    return () => window.clearInterval(id);
  }, [polling, intervalMs]);

  return { polling, setPolling };
}