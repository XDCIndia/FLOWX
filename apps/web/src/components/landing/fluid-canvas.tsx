'use client';

import { useEffect, useRef } from 'react';

/**
 * White-theme fluid ink canvas (ported from the original landing.html hero).
 * Loads /fluid-hero.js once and runs the WebGL sim on this canvas.
 * Falls back silently (plain white) when WebGL is unavailable.
 */
export default function FluidCanvas() {
  const ref = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    const canvas = ref.current;
    if (!canvas) return;
    let handle: { destroy: () => void } | null = null;
    let cancelled = false;

    const start = () => {
      const api = (window as unknown as { FluidHero?: { init: (c: HTMLCanvasElement) => { destroy: () => void } | null } }).FluidHero;
      if (cancelled || !api) return;
      try {
        handle = api.init(canvas);
      } catch {
        handle = null; // no WebGL — hero stays clean white
      }
    };

    if (document.querySelector('script[data-fluid-hero]')) {
      start();
    } else {
      const s = document.createElement('script');
      s.src = '/fluid-hero.js';
      s.async = true;
      s.dataset.fluidHero = '1';
      s.onload = start;
      document.body.appendChild(s);
    }

    return () => {
      cancelled = true;
      handle?.destroy();
    };
  }, []);

  return (
    <canvas
      ref={ref}
      aria-hidden="true"
      className="pointer-events-none absolute inset-0 h-full w-full"
    />
  );
}
