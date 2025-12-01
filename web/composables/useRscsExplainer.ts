import { inject, provide, ref, nextTick } from 'vue'
import type { Ref } from 'vue'

export type RscsExplainerStore = {
  visible: Ref<boolean>
  open: (triggerEl?: Element | null) => void
  close: () => void
  toggle: (triggerEl?: Element | null) => void
}

const KEY = Symbol('rscs-explainer')

export function provideRscsExplainer() {
  const visible = ref(false)
  let lastTrigger: Element | null = null

  const open = (triggerEl?: Element | null) => {
    lastTrigger = (triggerEl as Element) || (typeof document !== 'undefined' ? (document.activeElement as Element) : null)
    visible.value = true
  }

  const close = () => {
    visible.value = false
    if (lastTrigger && 'focus' in lastTrigger) {
      // Return focus to the triggering control after close
      nextTick(() => (lastTrigger as HTMLElement).focus())
    }
  }

  const toggle = (triggerEl?: Element | null) => {
    if (visible.value) close()
    else open(triggerEl)
  }

  const store: RscsExplainerStore = { visible, open, close, toggle }
  provide(KEY, store)
  return store
}

export function useRscsExplainer(): RscsExplainerStore {
  const store = inject<RscsExplainerStore>(KEY)
  if (!store) {
    // Fallback isolated store (useful in tests), but provider should exist in layout
    const visible = ref(false)
    return {
      visible,
      open: () => (visible.value = true),
      close: () => (visible.value = false),
      toggle: () => (visible.value = !visible.value),
    }
  }
  return store
}
