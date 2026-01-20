/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./ui/templates/**/*.html"],
  theme: {
    extend: {},
  },
  plugins: [require("daisyui")],
  daisyui: {
    themes: "all",
  },
}
