BootstrapButtons = Swal.mixin({
  customClass: { confirmButton: 'swal btn btn-primary' },
  buttonsStyling: false
});

valid = () => {
  var result = true
  Array.from(document.getElementsByTagName('input'))
    .forEach(i => { if (!i.checkValidity()) result = false })
  return result
}

post = (url, obj) => {
  return fetch(url, {
    method: 'post',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: new URLSearchParams(obj)
  })
    .catch(e => {
      return Promise.reject(BootstrapButtons.fire('Error', e, 'error'))
    })
    .then(resp => {
      if (resp.status == 401)
        return BootstrapButtons.fire('Error', 'Login status has changed. Please Re-login!', 'error')
          .then(() => window.location = '/')
      return resp
    })
}

confirm = type => {
  return Swal.fire({
    title: 'Are you sure?',
    text: 'This ' + type + ' will be deleted permanently.',
    icon: 'warning',
    confirmButtonText: 'Delete',
    showCancelButton: true,
    focusCancel: true,
    customClass: {
      confirmButton: 'swal btn btn-danger',
      cancelButton: 'swal btn btn-primary'
    },
    buttonsStyling: false
  }).then(confirm => { return confirm.isConfirmed })
}

const app = Vue.createApp({
  data() { return { user: document.getElementById('app').dataset.user } },
  computed: {
    loading() { return this.$store.state.loading },
    sidebar() { return this.$store.state.sidebar },
    component() { return this.$store.state.component }
  },
  methods: { setting: function () { this.$store.commit('goto', 'setting') } }
})

const store = Vuex.createStore({
  state() {
    return {
      component: 'showBookmark',
      sidebar: false,
      loading: false,
      categories: [],
      bookmarks: [],
      category: {},
      bookmark: {},
      editCategory: {}
    }
  },
  mutations: {
    goto(state, component) { state.component = component },
    loading(state, status) { state.loading = status },
    ready(state) { state.sidebar = true },
    categories(state) {
      state.sidebar = false
      state.loading = true
      post('/category/get')
        .then(response => response.json())
        .then(categories => {
          state.categories = categories
          if (state.category.count == undefined)
            state.category = {
              id: -1,
              name: 'All Bookmarks',
              count: categories.reduce((total, i) => total + i.count, 0),
              start: 0
            }
          state.loading = false
          state.sidebar = true
        })
    },
    bookmarks(state, payload) {
      state.loading = true
      if (payload.more) {
        state.category.start += 30
        var params = { category: state.category.id, start: state.category.start }
      } else var params = { category: payload.id }
      post('/bookmark/get', params)
        .then(resp => {
          if (!resp.ok) resp.text().then(err => {
            return BootstrapButtons.fire('Error', err, 'error')
          })
          else resp.json().then(bookmarks => {
            if (payload.more)
              state.bookmarks = state.bookmarks.concat(bookmarks)
            else {
              state.bookmarks = bookmarks
            }
          })
        }).then(() => state.loading = false)
    },
    category(state, category) { state.category = category },
    bookmark(state, bookmark) { state.bookmark = bookmark },
    editCategory(state, category) { state.editCategory = category },
    renCategory(state, name) { state.bookmarks.forEach(i => i.category = name) },
    delBookmarks(state, bookmark) {
      state.categories.forEach(i => { if (i.name == bookmark.category) i.count-- })
      state.bookmarks = state.bookmarks.filter(i => i.id != bookmark.id)
    }
  }
})
app.use(store)

app.mixin({
  methods: {
    goback: function (reload) {
      if (reload)
        this.$store.commit('categories')
      this.$store.commit('goto', 'showBookmark')
    }
  }
})

app.component('login', login)
app.component('setting', setting)
app.component('sidebar', sidebar)
app.component('showBookmark', showBookmarks)
app.component('category', category)
app.component('bookmark', bookmark)

app.mount('#app')
