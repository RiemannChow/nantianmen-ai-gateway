export interface PromptTemplate {
  id: string
  name: string
  tags: string[]
  content: string
  variables: string[]
  updated_at: string
}

export interface PromptManagementHandle {
  save: () => Promise<boolean | string>
}
