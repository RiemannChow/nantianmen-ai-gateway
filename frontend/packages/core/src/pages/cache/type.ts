export interface SemanticCacheConfig {
  enabled: boolean
  ttl_seconds: number
  embedding_endpoint: string
  embedding_key: string
}

export interface SemanticCacheStats {
  hits: number
  misses: number
  hit_rate: number
  entries: number
}

export interface SemanticCacheEntry {
  key: string
  model: string
  tenant: string
  prompt_hash: string
  response: string
  token_count: number
  created_at: string
}
