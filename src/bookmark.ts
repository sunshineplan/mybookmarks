import { writable, get } from 'svelte/store'
import type { Writable } from 'svelte/store'
import { total, username } from "./stores";
import { fire, post } from './misc'

export interface Category {
  category: string
  count: number
}

export interface Bookmark {
  id: string
  category: string
  bookmark: string
  url: string
  seq: number
}

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

const more = async (init?: boolean) => {
  const currentCategory = get(category)
  const currentBookmarks = get(bookmarks)
  const now = currentCategory.category == 'All Bookmarks'
    ? currentBookmarks.length
    : currentBookmarks.filter(b => b.category == currentCategory.category).length
  const uncategorized = get(total) - get(categories).reduce((a, b) => a + b.count, 0)
  const goal = currentCategory.category == 'All Bookmarks'
    ? get(total)
    : currentCategory.category
      ? currentCategory.count
      : uncategorized
  if (now >= (init ? Math.min(30, goal) : goal)) return
  const resp = await post('/bookmark/get', { start: currentBookmarks.length })
  if (resp.ok) {
    const moreBookmarks = currentBookmarks.concat(await resp.json())
    moreBookmarks.sort((a, b) => a.seq - b.seq)
    bookmarks.set(moreBookmarks)
    if (currentCategory.category == 'All Bookmarks') return
    const moreCount = moreBookmarks.filter(b => b.category == currentCategory.category).length
    if (moreCount < now + 15)
      if (currentCategory.category) { if (moreCount < currentCategory.count) await more(init) }
      else if (moreCount < uncategorized) await more(init)
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
