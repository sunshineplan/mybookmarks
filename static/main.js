const app = Vue.createApp({
  delimiters: ['{%', '%}'],
  data() {
    return {
      user: document.getElementById('bookmark').dataset.user,
      content: 'showBookmark',
      current: { id: -1, category: 'All Bookmarks' },
      siderbar: false,
      loading: false,
      active: -1,
      category: {},
      bookmark: {}
    }
  },
  computed: {
    prop: function () {
      if (this.content == 'showBookmark')
        return { current: this.current }
      else if (this.content == 'category')
        return { category: this.category }
      else if (this.content == 'bookmark')
        return {
          bookmark: this.bookmark,
          categories: this.$refs.categories.category.categories
        }
    }
  },
  methods: {
    setting: function () {
      this.content = 'setting'
      this.active = null
    }
  }
})

app.component('login', login)
app.component('setting', setting)
app.component('sidebar', sidebar)
app.component('showBookmark', showBookmarks)
app.component('category', category)
app.component('bookmark', bookmark)

app.mount('#bookmark')
