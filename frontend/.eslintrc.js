module.exports = {
  extends: [
    'plugin:@typescript-eslint/recommended',
    'plugin:react/recommended', // For React projects (remove if not using React)
    'prettier', // Disables ESLint rules that conflict with Prettier
  ],
}
