import { Dexie } from 'dexie'
import { getCookie } from 'typescript-cookie'
import { fire, loading, post } from './misc.svelte'

const db = new Dexie('bookmark')
db.version(1).stores({
  categories: 'category',
  bookmarks: 'id,category'
})

class MyBookmarks {
  username = $state('')
  component = $state('show')
  category = $state<Category>({})
  bookmark = $state<Bookmark>({} as Bookmark)
  categories = $state<Category[]>([])
  bookmarks = $state<Bookmark[]>([])
  controller = $state(new AbortController())
  async clear() {
    await db.table('categories').clear()
    await db.table('bookmarks').clear()
  }
  async reset() {
    this.username = ''
    this.category = {}
    this.bookmark = {} as Bookmark
    this.categories = []
    this.bookmarks = []
    await this.clear()
  }
  async init() {
    loading.start()
    let resp: Response
    try {
      resp = await fetch('/info')
    } catch (e) {
      console.error(e)
      resp = new Response(null, { "status": 500 })
    }
    loading.end()
    if (resp.ok) {
      const username = await resp.text()
      if (username) {
        await this.getCategories()
        await this.getBookmarks()
        this.username = username
      } else await this.reset()
    } else if (resp.status == 409) {
      await this.clear()
      await this.init()
    } else await this.reset()
  }
  async getCategories() {
    await db.table<Category>('categories').filter(i => i.count == 0).delete()
    const array = await db.table('categories').toArray()
    if (array.length) this.categories = array
    else await this.#fetchCategories()
  }
  async #fetchCategories() {
    const resp = await post('/category/get')
    if (resp.ok) {
      const res = await resp.json()
      this.categories = res
      await db.table('categories').bulkAdd(res)
    } else await fire('Fatal', await resp.text(), 'error')
  }
  async #count(category?: Category) {
    let n = 0
    if (category === undefined || category.category === undefined) await db.table('categories').each(i => n += i.count)
    else await db.table<Category>('categories').where('category').equals(category.category).first(i => n = i ? i.count || 0 : 0)
    return n
  }
  async addCategory(category: Category) {
    await db.table('categories').add(category)
    const array = [...this.categories]
    if (array.slice(-1)[0].category == '') {
      array.splice(array.length - 1, 0, category)
      this.categories = array
    } else this.categories = [...array, category]
  }
  async editCategory(category: Category, name: string) {
    this.controller.abort()
    const resp = await post('/category/edit', { old: category.category, new: name })
    let msg = ''
    if (resp.ok) {
      const res = await resp.json()
      if (res.status) {
        await db.table('categories').update(category.category, { category: name })
        await db.table('bookmarks').where('category').equals(category.category || '').modify({ category: name })
        this.categories = await db.table('categories').toArray()
        this.subscribe(true)
        return
      } else msg = res.message
    } else msg = await resp.text()
    await fire('Fatal', msg, 'error')
    this.subscribe(true)
  }
  async deleteCategory(category: Category) {
    this.controller.abort()
    const resp = await post('/category/delete', { category })
    if (resp.ok) {
      await db.table('categories').where('category').equals(category.category || '').delete()
      const n = await db.table<Category>('categories').where('category').equals('').modify(i => {
        if (i && i.count && category && category.count) {
          i.count += category.count
        }
      })
      if (!n) await db.table('categories').add({ category: '', count: category.count })
      await db.table('bookmarks').where('category').equals(category.category || '').modify({ category: '' })
      this.categories = await db.table('categories').toArray()
    } else await fire('Fatal', await resp.text(), 'error')
    this.subscribe(true)
  }
  async getBookmarks(category?: Category, more?: number, goal?: number) {
    const res = await this.#loadBookmarks(category)
    const count = await this.#count(category)
    if (!goal)
      if (more) goal = Math.min(res.length + more, count)
      else goal = Math.min(30, count)
    if (res.length >= goal) return
    await this.#fetchBookmarks(await db.table('bookmarks').count())
    await this.getBookmarks(category, more, goal)
  }
  async #loadBookmarks(category?: Category) {
    if (category === undefined || category.category === undefined)
      this.bookmarks = await db.table('bookmarks').toCollection().sortBy('seq')
    else this.bookmarks = await db.table('bookmarks').where('category').equals(category.category).sortBy('seq')
    return this.bookmarks
  }
  async #fetchBookmarks(start: number) {
    const resp = await post('/bookmark/get', { start })
    if (resp.ok) await db.table('bookmarks').bulkAdd(await resp.json())
    else await fire('Fatal', await resp.text(), 'error')
  }
  async saveBookmark(bookmark: Bookmark) {
    let resp: Response | undefined = undefined
    if (bookmark.id) resp = await post('/bookmark/edit/' + bookmark.id, bookmark)
    else resp = await post('/bookmark/add', bookmark)
    if (resp.ok) {
      const res = await resp.json()
      if (res.status == 1) {
        if (bookmark.id) await db.table('bookmarks').update(bookmark.id, bookmark)
        else {
          bookmark.id = res.id
          bookmark.seq = res.seq
          await db.table('bookmarks').add(bookmark)
        }
        await db.table('categories').clear()
        await this.#fetchCategories()
        await this.getBookmarks({ category: bookmark.category })
      } else {
        await fire('Error', res.message, 'error')
        return <number>res.error
      }
    } else await fire('Fatal', await resp.text(), 'error')
    return 0
  }
  async deleteBookmark(bookmark: Bookmark) {
    const resp = await post('/bookmark/delete/' + bookmark.id)
    if (resp.ok) {
      await db.table('bookmarks').where('id').equals(bookmark.id).delete()
      await db.table('categories').where('category').equals(bookmark.category).modify((i: Category) => i && i.count && i.count--)
      const category = await db.table<Category>('categories').get({ 'category': bookmark.category })
      if (!category?.count) await db.table('categories').where('category').equals(bookmark.category).delete()
      this.categories = await db.table('categories').toArray()
      await this.getBookmarks({ category: bookmark.category })
      if (!this.bookmarks.length) await this.getBookmarks()
    } else await fire('Fatal', await resp.text(), 'error')
  }
  async swap(a: Bookmark, b: Bookmark) {
    this.controller.abort()
    const resp = await post('/reorder', { orig: a.id, dest: b.id })
    if (resp.ok) {
      if ((await resp.text()) == '1') {
        const array = await db.table('bookmarks').toCollection().sortBy('seq')
        const seq = b.seq
        if (a.seq > b.seq) array.forEach(i => { if (i.seq >= b.seq && i.seq < a.seq) i.seq++ })
        else array.forEach(i => { if (i.seq > a.seq && i.seq <= b.seq) i.seq-- })
        array.forEach(i => { if (i.id === a.id) i.seq = seq })
        array.sort((a, b) => a.seq - b.seq)
        await db.table('bookmarks').bulkPut(array)
      } else await fire('Fatal', 'Failed to reorder.', 'error')
    } else await fire('Fatal', await resp.text(), 'error')
    this.subscribe(true)
  }
  async subscribe(reset?: boolean) {
    if (reset)
      this.controller = new AbortController()
    let resp: Response
    try {
      resp = await fetch('/poll', { signal: this.controller.signal })
    } catch (e) {
      if (e instanceof DOMException && e.name === 'AbortError') return
      console.error(e)
      resp = new Response(null, { status: 500 })
    }
    if (resp.ok) {
      const last = await resp.text()
      if (last && getCookie('last') != last) {
        await this.init()
      }
      await this.subscribe()
    } else if (resp.status == 401) {
      await this.init()
    } else {
      await new Promise((sleep) => setTimeout(sleep, 30000))
      await this.subscribe()
    }
  }
}
export const mybookmarks = new MyBookmarks
