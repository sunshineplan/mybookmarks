import { writable } from 'svelte/store'

export const username = writable('')
export const component = writable('show')

const createShowSidebar = () => {
  const { subscribe, set, update } = writable(false)
  return {
    subscribe,
    toggle: () => update(status => !status),
    close: () => set(false)
  }
}
export const showSidebar = createShowSidebar()

const createLoading = () => {
  const { subscribe, update } = writable(0)
  return {
    subscribe,
    start: () => update(n => n + 1),
    end: () => update(n => n - 1)
  }
}
export const loading = createLoading()
