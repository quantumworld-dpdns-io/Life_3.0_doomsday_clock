import type { Config } from "tailwindcss";

export default {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        bg: "var(--bg)",
        "accent-red": "var(--accent-red)",
        "accent-green": "var(--accent-green)",
        "text-primary": "var(--text-primary)",
      },
      fontFamily: {
        mono: "var(--font-mono)",
        display: "var(--font-display)",
      },
    },
  },
} satisfies Config;
