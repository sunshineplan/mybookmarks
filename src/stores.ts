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
    more: async () => {
      console.log('more')
      let start: number
      if (get(category).id === -1) {
        start = get(bookmarks).length
        if (start >= get(total)) return
      } else if (get(category).id === 0) {
        start = get(bookmarks).filter((b) => b.category == '').length
        if (
          start >=
          get(total) - get(categories).reduce((a, b) => a + b.count, 0)
        )
          return
      } else {
        start = get(bookmarks).filter((b) => b.category == get(category).category).length
        if (length >= get(category).count) return
      }
      loading.update(n => n + 1)
      const resp = await post('/bookmark/get', {
        category: get(category).id,
        start,
      })
      loading.update(n => n - 1)
      if (resp.ok) {
        const current = get(bookmarks).concat(await resp.json())
        current.sort((a, b) => a.seq - b.seq)
        bookmarks.set(current)
      } else await fire('Error', await resp.text(), 'error')
    }
  }
}
export const bookmarks = createBookmarks()
