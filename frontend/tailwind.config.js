/** @type {import('tailwindcss').Config} */
module.exports = {

  content: [
    "./templates/*.tmpl",
    "./static/**/*.html",
  ],

  theme: {},

  plugins: [require("@tailwindcss/typography"), require("daisyui")],

  daisyui: {
    themes: [
      {
        mytheme: {
          "primary": "#1DDDDD",
          "secondary": "#1D1E20",
          "accent": "#DADADB",
          "neutral": "#1D1E20",
          "base-100": "#1D1E20",
          "info": "#3ABFF8",
          "success": "#36D399",
          "warning": "#FBBD23",
          "error": "#F87272",
        },
      },
    ],
  },
}


