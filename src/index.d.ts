declare interface Window {
  universal: string
  pubkey: string
}

declare interface Category {
  category: string
  count: number
}

declare interface Bookmark {
  id: string
  category: string
  bookmark: string
  url: string
  seq: number
}
