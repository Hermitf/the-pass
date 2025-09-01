/// <reference types="vite/client" />

declare interface ImportMetaEnv {
    readonly MODE: string
//   readonly VITE_API_BASE_URL: string
//   readonly VITE_APP_ENV: string
//   readonly VITE_APP_NAME: string
//   readonly VITE_APP_VERSION: string
  // 更多环境变量...
}

declare interface ImportMeta {
  readonly env: ImportMetaEnv
}
