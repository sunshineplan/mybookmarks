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
export const category: Writable<Category> = writable({ id: -1, category: 'All Bookmarks', count: 0 })
export const bookmark: Writable<Bookmark> = writable({} as Bookmark)
export const categories: Writable<Category[]> = writable([])
export const showSidebar = writable(false)
export const loading = writable(0)

const createBookmarks = () => {
  const { subscribe, set } = writable([] as Bookmark[])
  return {
    subscribe,
    set,
    more: async (init?: boolean) => {
      const currentCategory = get(category)
      const currentBookmarks = get(bookmarks)
      let start: number, goal: number
      switch (currentCategory.id) {
        case -1:
          start = currentBookmarks.length
          goal = get(total)
          break
        case 0:
          start = currentBookmarks.filter(b => b.category == '').length
          goal = get(total) - get(categories).reduce((a, b) => a + b.count, 0)
          break
        default:
          start = currentBookmarks.filter(b => b.category == currentCategory.category).length
          goal = currentCategory.count
      }
      if (start >= (init ? Math.min(30, goal) : goal)) return
      loading.update(n => n + 1)
      const resp = await post('/bookmark/get', { category: currentCategory.id, start })
      loading.update(n => n - 1)
      if (resp.ok) {
        const moreBookmarks = currentBookmarks.concat(await resp.json())
        moreBookmarks.sort((a, b) => a.seq - b.seq)
        bookmarks.set(moreBookmarks)
      } else await fire('Error', await resp.text(), 'error')
    }
  }
}
export const bookmarks = createBookmarks()
