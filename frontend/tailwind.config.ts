import type { Config } from "tailwindcss";

const config: Config = {
  content: [
    "./app/**/*.{js,ts,jsx,tsx,mdx}",
    "./components/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      colors: {
        bg: "#f8fafc",
        panel: "#ffffff",
        text: "#0f172a",
        muted: "#64748b",
        accent: "#0ea5e9",
        positive: "#10b981",
        negative: "#ef4444",
        teal: {
          50: '#f0fdfa',
          100: '#ccfbf1',
          500: '#14b8a6',
          600: '#0d9488',
          700: '#0f766e',
          900: '#134e4a',
        }
      },
      boxShadow: {
        panel: "0 20px 40px rgba(15, 23, 42, 0.04)",
        glass: "0 8px 32px 0 rgba(31, 38, 135, 0.05)",
      },
    },
  },
  plugins: [],
};

export default config;
