import { computed } from 'vue'

import { pickDefaultScopeContext, type DataScope, type ScopeCreateContext } from '@/utils/dataScope'
import { useAuthStore } from '@/stores/auth'

const NO_SCOPE_MSG =
  'No data scope is assigned to your account. Ask an administrator to assign institution access (and department or team if needed).'

export function useCreateScopeContext() {
  const auth = useAuthStore()
  const context = computed(() => pickDefaultScopeContext(auth.scopes as DataScope[]))

  function requireContext(): ScopeCreateContext {
    const ctx = context.value
    if (!ctx) {
      throw new Error(NO_SCOPE_MSG)
    }
    return ctx
  }

  return { context, requireContext, noScopeMessage: NO_SCOPE_MSG }
}
