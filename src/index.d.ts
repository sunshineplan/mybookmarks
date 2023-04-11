interface Window {
  universal: string
  pubkey: string
}

interface Category {
  category?: string
  count?: number
}

interface Bookmark {
  id: string
  category: string
  bookmark: string
  url: string
  seq: number
}
