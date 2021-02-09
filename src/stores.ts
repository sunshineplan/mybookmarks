import { Writable, writable, get } from 'svelte/store'
import { fire, post } from './misc'

export interface Category {
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
  const { subscribe, set } = writable({ category: 'All Bookmarks', count: 0 } as Category)
  return {
    subscribe,
    set,
    reset: () => set({ category: 'All Bookmarks', count: 0 }),
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
  const currentCategory = get(category)
  const currentBookmarks = get(bookmarks)
  const now = currentCategory.category == 'All Bookmarks'
    ? currentBookmarks.length
    : currentBookmarks.filter(b => b.category == currentCategory.category).length
  const goal = currentCategory.category == 'All Bookmarks'
    ? get(total)
    : currentCategory.category
      ? currentCategory.count
      : get(total) - get(categories).reduce((a, b) => a + b.count, 0)
  if (now >= (init ? Math.min(30, goal) : goal)) return
  loading.start()
  const resp = await post('/bookmark/get', { start: currentBookmarks.length })
  loading.end()
  if (resp.ok) {
    const moreBookmarks = currentBookmarks.concat(await resp.json())
    moreBookmarks.sort((a, b) => a.seq - b.seq)
    bookmarks.set(moreBookmarks)
    if (currentCategory.category == 'All Bookmarks') return
    const moreCount = moreBookmarks.filter(b => b.category == currentCategory.category).length
    if (moreCount < now + 15)
      if (currentCategory.category && moreCount < currentCategory.count) await more(init)
      else if (moreCount < goal - get(categories).reduce((a, b) => a + b.count, 0)) await more(init)
  } else await fire('Error', await resp.text(), 'error')
}

const createBookmarks = () => {
  const { subscribe, set } = writable([] as Bookmark[])
  return { subscribe, set, more }
}
export const bookmarks = createBookmarks()

export const reset = () => {
  username.set('')
  bookmark.set({} as Bookmark)
  categories.set([])
  category.reset()
  bookmarks.set([])
}
