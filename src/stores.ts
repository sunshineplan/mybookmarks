import { Writable, writable, get } from 'svelte/store'
import { fire, post } from './misc'

export interface Category {
  id: number
  category: string
  count: number
}

export interface Bookmark {
  id: number
  category: string
  bookmark: string
  url: string
  seq: number
}

export const username = writable('')
export const total = writable(0)
export const component = writable('show')
export const bookmark: Writable<Bookmark> = writable({} as Bookmark)
export const categories: Writable<Category[]> = writable([])

const createCategory = () => {
  const { subscribe, set } = writable({ id: -1, category: 'All Bookmarks', count: 0 } as Category)
  return {
    subscribe,
    set,
    reset: () => set({ id: -1, category: 'All Bookmarks', count: 0 }),
  }
}
export const category = createCategory()

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

const more = async (init?: boolean) => {
  const currentBookmarks = get(bookmarks)
  const start = currentBookmarks.length
  const goal = get(total)
  if (start >= (init ? Math.min(30, goal) : goal)) return
  loading.start()
  const resp = await post('/bookmark/get', { start })
  loading.end()
  if (resp.ok) {
    const moreBookmarks = currentBookmarks.concat(await resp.json())
    moreBookmarks.sort((a, b) => a.seq - b.seq)
    bookmarks.set(moreBookmarks)
    const currentCategory = get(category)
    if (currentCategory.id == -1) return
    const moreCount = moreBookmarks.filter(b => b.category == currentCategory.category).length
    if (moreCount < currentBookmarks.filter(b => b.category == currentCategory.category).length + 15)
      if (currentCategory.id && moreCount < currentCategory.count) await more()
      else if (moreCount < goal - get(categories).reduce((a, b) => a + b.count, 0)) await more()
  } else await fire('Error', await resp.text(), 'error')
}

const createBookmarks = () => {
  const { subscribe, set } = writable([] as Bookmark[])
  return { subscribe, set, more }
}
export const bookmarks = createBookmarks()
