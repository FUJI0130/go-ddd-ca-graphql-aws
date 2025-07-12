/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_GRAPHQL_API_URL: string
  readonly VITE_APP_NAME: string
  readonly VITE_APP_VERSION: string
  readonly VITE_JWT_LOCAL_STORAGE_KEY: string
  readonly VITE_REFRESH_TOKEN_LOCAL_STORAGE_KEY: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}