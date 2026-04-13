/// <reference types="vite/client" />

interface ImportMetaEnv {
  /** Optional absolute API origin; leave unset to use same-origin `/api` (Docker nginx or Vite proxy). */
  readonly VITE_API_BASE_URL?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<object, object, any>
  export default component
}
