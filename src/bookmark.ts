import { writable, get } from 'svelte/store'
import { Dexie } from 'dexie'
import { username } from './stores'
import { fire, post } from './misc'

const db = new Dexie('bookmark')
db.version(1).stores({
  categories: 'category',
  bookmarks: 'id,category'
})

export const category = writable(<Category>{})
export const bookmark = writable(<Bookmark>{})

const createCategories = () => {
  const { subscribe, set } = writable(<Category[]>[])
  return {
    subscribe,
    set,
    clear: async () => await db.table('categories').clear(),
    get: async () => {
      await db.table<Category>('categories').filter(i => i.count == 0).delete()
      const array = await db.table('categories').toArray()
      if (array.length) categories.set(array)
      else await categories.fetch()
    },
    fetch: async () => {
      const resp = await post('/category/get')
      if (resp.ok) {
        const res = await resp.json()
        categories.set(res)
        await db.table('categories').bulkAdd(res)
      } else await fire('Fatal', await resp.text(), 'error')
    },
    count: async (category?: Category) => {
      let n = 0
      if (category === undefined || category.category === undefined) await db.table('categories').each(i => n += i.count)
      else await db.table<Category>('categories').where('category').equals(category.category).first(i => n = i.count)
      return n
    },
    add: async (category: Category) => {
      await db.table('categories').add(category)
      const array = get(categories)
      if (array.slice(-1)[0].category == '') {
        array.splice(array.length - 1, 0, category)
        categories.set(array)
      } else categories.set([...array, category])
    },
    edit: async (category: Category, name: string) => {
      const resp = await post('/category/edit', { old: category.category, new: name })
      let msg = ''
      if (resp.ok) {
        const res = await resp.json()
        if (res.status) {
          await db.table('categories').update(category.category, { category: name })
          await db.table('bookmarks').where('category').equals(category.category).modify({ category: name })
          categories.set(await db.table('categories').toArray())
          return
        } else msg = res.message
      } else msg = await resp.text()
      await fire('Fatal', msg, 'error')
    },
    delete: async (category: Category) => {
      const resp = await post('/category/delete', { category })
      if (resp.ok) {
        await db.table('categories').where('category').equals(category.category).delete()
        const n = await db.table<Category>('categories').where('category').equals('').modify(i => i.count += category.count)
        if (!n) await db.table('categories').add({ category: '', count: category.count })
        await db.table('bookmarks').where('category').equals(category.category).modify({ category: '' })
        categories.set(await db.table('categories').toArray())
      } else await fire('Fatal', await resp.text(), 'error')
    }
  }
}
export const categories = createCategories()

const createBookmarks = () => {
  const { subscribe, set } = writable(<Bookmark[]>[])
  return {
    subscribe,
    set,
    clear: async () => await db.table('bookmarks').clear(),
    get: async (category?: Category, more?: number, goal?: number) => {
      await bookmarks.load(category)
      const res = get(bookmarks)
      const count = await categories.count(category)
      if (!goal)
        if (more) goal = Math.min(res.length + more, count)
        else goal = Math.min(30, count)
      if (res.length >= goal) return
      await bookmarks.fetch(await db.table('bookmarks').count())
      await bookmarks.get(category, more, goal)
    },
    load: async (category?: Category) => {
      if (category === undefined || category.category === undefined)
        bookmarks.set(await db.table('bookmarks').toCollection().sortBy('seq'))
      else bookmarks.set(await db.table('bookmarks').where('category').equals(category.category).sortBy('seq'))
    },
    fetch: async (start: number) => {
      const resp = await post('/bookmark/get', { start })
      if (resp.ok) await db.table('bookmarks').bulkAdd(await resp.json())
      else await fire('Fatal', await resp.text(), 'error')
    },
    save: async (bookmark: Bookmark) => {
      let resp: Response = undefined
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
          await categories.clear()
          await categories.fetch()
          await bookmarks.get({ category: bookmark.category })
        } else {
          await fire('Error', res.message, 'error')
          return <number>res.error
        }
      } else await fire('Fatal', await resp.text(), 'error')
      return 0
    },
    delete: async (bookmark: Bookmark) => {
      const resp = await post('/bookmark/delete/' + bookmark.id)
      if (resp.ok) {
        await db.table('bookmarks').where('id').equals(bookmark.id).delete()
        await db.table('categories').where('category').equals(bookmark.category).modify((i: Category) => i.count--)
        const category = await db.table<Category>('categories').get({ 'category': bookmark.category })
        if (!category.count) await db.table('categories').where('category').equals(bookmark.category).delete()
        categories.set(await db.table('categories').toArray())
        await bookmarks.get({ category: bookmark.category })
        if (!get(bookmarks).length) await bookmarks.get()
      } else await fire('Fatal', await resp.text(), 'error')
    },
    swap: async (a: Bookmark, b: Bookmark) => {
      const resp = await post('/reorder', { orig: a.id, dest: b.id })
      if (resp.ok) {
        if ((await resp.text()) == '1') {
          const array = await db.table('bookmarks').toCollection().sortBy('seq')
          if (a.seq > b.seq) array.forEach(i => { if (i.seq >= b.seq && i.seq < a.seq) i.seq++ })
          else array.forEach(i => { if (i.seq > a.seq && i.seq <= b.seq) i.seq-- })
          array.forEach(i => { if (i.id === a.id) i.seq = b.seq })
          array.sort((a, b) => a.seq - b.seq)
          await db.table('bookmarks').bulkPut(array)
        } else await fire('Fatal', 'Failed to reorder.', 'error')
      } else await fire('Fatal', await resp.text(), 'error')
    }
  }
}
export const bookmarks = createBookmarks()

export const init = async () => {
  const resp = await fetch('/info')
  if (resp.ok) {
    const info = await resp.json()
    if (Object.keys(info).length) {
      username.set(info.username)
      await categories.get()
      await bookmarks.get()
    } else await reset()
  } else if (resp.status == 409) {
    await categories.clear()
    await bookmarks.clear()
    await init()
  } else await reset()
}

const reset = async () => {
  username.set('')
  category.set(<Category>{})
  bookmark.set(<Bookmark>{})
  categories.set([])
  bookmarks.set([])
  await categories.clear()
  await bookmarks.clear()
}
