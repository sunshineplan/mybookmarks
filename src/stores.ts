import { Writable, writable } from 'svelte/store'

export interface Category {
  id: number
  category: string
  start?: number
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
export const component = writable('show')
export const category: Writable<Category> = writable({ id: -1, category: 'All Bookmarks', count: 0 })
export const bookmark: Writable<Bookmark> = writable({} as Bookmark)
export const categories: Writable<Category[]> = writable([])
export const bookmarks: Writable<Bookmark[]> = writable([])
export const showSidebar = writable(false)
export const loading = writable(0)
