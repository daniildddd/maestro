import { useEffect, useLayoutEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";

export interface MenuItem {
  label: string;
  onSelect: () => void;
  danger?: boolean;
  disabled?: boolean;
}

export function ActionMenu({ items, label = "Actions" }: { items: MenuItem[]; label?: string }) {
  const [open, setOpen] = useState(false);
  const [pos, setPos] = useState<{ top: number; left: number } | null>(null);
  const rootRef = useRef<HTMLDivElement | null>(null);
  const popRef = useRef<HTMLDivElement | null>(null);

  // position after paint but before the browser shows it: flip above the button
  // when there is not enough room below (last table rows would clip otherwise)
  useLayoutEffect(() => {
    if (!open || !rootRef.current || !popRef.current) return;
    const btn = rootRef.current.getBoundingClientRect();
    const pop = popRef.current.getBoundingClientRect();
    const spaceBelow = window.innerHeight - btn.bottom;
    const top = spaceBelow < pop.height + 8 ? Math.max(8, btn.top - pop.height - 4) : btn.bottom + 4;
    const left = Math.max(8, Math.min(btn.right - pop.width, window.innerWidth - pop.width - 8));
    setPos({ top, left });
  }, [open]);

  useEffect(() => {
    if (!open) return;
    const onDown = (e: MouseEvent) => {
      const target = e.target as Node;
      if (rootRef.current?.contains(target) || popRef.current?.contains(target)) return;
      setOpen(false);
    };
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") setOpen(false);
    };
    const onScroll = () => setOpen(false);
    document.addEventListener("mousedown", onDown);
    document.addEventListener("keydown", onKey);
    document.addEventListener("scroll", onScroll, true);
    return () => {
      document.removeEventListener("mousedown", onDown);
      document.removeEventListener("keydown", onKey);
      document.removeEventListener("scroll", onScroll, true);
    };
  }, [open]);

  return (
    <div className="action-menu" ref={rootRef}>
      <button
        type="button"
        className="icon-btn"
        aria-haspopup="menu"
        aria-expanded={open}
        aria-label={label}
        onClick={() => setOpen((v) => !v)}
      >
        ⋯
      </button>
      {open &&
        createPortal(
          <div
            ref={popRef}
            className="menu-pop"
            role="menu"
            style={
              pos
                ? { position: "fixed", top: pos.top, left: pos.left }
                : { position: "fixed", top: -9999, left: -9999, visibility: "hidden" }
            }
          >
            {items.map((it) => (
              <button
                key={it.label}
                type="button"
                role="menuitem"
                className={"menu-item" + (it.danger ? " danger" : "")}
                disabled={it.disabled}
                onClick={() => {
                  setOpen(false);
                  it.onSelect();
                }}
              >
                {it.label}
              </button>
            ))}
          </div>,
          document.body,
        )}
    </div>
  );
}