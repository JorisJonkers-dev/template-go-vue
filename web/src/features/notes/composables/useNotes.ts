import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import { computed } from 'vue'
import {
  createNoteMutation,
  listNotesOptions,
  listNotesQueryKey,
} from '../../../infrastructure/api/@tanstack/vue-query.gen'
import type { Note, Problem } from '../../../infrastructure/api/types.gen'

function isProblem(error: unknown): error is Problem {
  return typeof error === 'object' && error !== null && 'title' in error && 'status' in error
}

/** What to tell the user about a failed request: the problem's detail when the server sent one. */
export function describeError(error: unknown): string {
  if (isProblem(error)) return error.detail ?? error.title
  return 'Something went wrong. Try again.'
}

/** Server state for the notes feature. Components read notes through this and never call the API. */
export function useNotes() {
  const queryClient = useQueryClient()
  const list = useQuery(listNotesOptions())
  const create = useMutation(createNoteMutation())

  return {
    notes: computed<Note[]>(() => list.data.value?.items ?? []),
    isLoading: list.isPending,
    loadError: computed(() => (list.error.value ? describeError(list.error.value) : null)),
    isCreating: create.isPending,
    createError: computed(() => (create.error.value ? describeError(create.error.value) : null)),
    create: (text: string, onCreated: () => void) => {
      create.mutate(
        { body: { text } },
        {
          onSuccess: () => {
            onCreated()
            void queryClient.invalidateQueries({ queryKey: listNotesQueryKey() })
          },
        },
      )
    },
  }
}
