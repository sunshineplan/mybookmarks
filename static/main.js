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

post = (url, data) => {
  return fetch(url, {
    method: 'post',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data)
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
  },
  methods: { setting: function () { this.$router.push('/setting') } }
})

const store = Vuex.createStore({
  state() {
    return {
      sidebar: false,
      loading: 0,
      categories: [],
      bookmarks: [],
      category: {},
      bookmark: {}
    }
  },
  mutations: {
    startLoading(state) { state.loading += 1 },
    stopLoading(state) { state.loading -= 1 },
    setSidebar(state, status) { state.sidebar = status },
    categories(state, categories) { state.categories = categories },
    bookmarks(state, bookmarks) { state.bookmarks = bookmarks },
    category(state, category) { state.category = category },
    more(state) { state.category.start += 30 },
    bookmark(state, bookmark) { state.bookmark = bookmark }
  },
  actions: {
    categories({ commit, state }) {
      commit('setSidebar', false)
      commit('startLoading')
      return post('/category/get')
        .then(response => response.json())
        .then(categories => {
          commit('categories', categories)
          if (state.category.count == undefined)
            commit('category', {
              id: -1,
              name: 'All Bookmarks',
              count: categories.reduce((total, i) => total + i.count, 0),
              start: 0
            })
          commit('stopLoading')
          commit('setSidebar', true)
        })
    },
    bookmarks({ commit, state }, payload) {
      commit('startLoading')
      if (payload.more) {
        commit('more')
        var params = { category: state.category.id, start: state.category.start }
      } else var params = { category: payload.id }
      return post('/bookmark/get', params)
        .then(resp => {
          if (!resp.ok) resp.text().then(err => {
            return BootstrapButtons.fire('Error', err, 'error')
          })
          else resp.json().then(bookmarks => {
            if (payload.more)
              commit('bookmarks', state.bookmarks.concat(bookmarks))
            else commit('bookmarks', bookmarks)
          })
        }).then(() => commit('stopLoading'))
    },
    addCategory({ dispatch, commit, state }, name) {
      return dispatch('categories')
        .then(() => { return state.categories.filter(i => i.name == name) })
        .then(category => {
          if (category.length) {
            commit('category', category[0])
            commit('bookmarks', [])
          }
        })
    },
    editCategory({ dispatch, commit, state }, name) {
      return dispatch('categories').then(() => {
        commit('category', {
          id: state.category.id,
          name,
          count: state.category.count,
          start: state.category.start
        })
        var bookmarks = state.bookmarks
        if (bookmarks)
          bookmarks.forEach(i => i.category = name)
        commit('bookmarks', bookmarks)
      })
    },
    delBookmarks({ commit, state }, bookmark) {
      var categories = state.categories
      categories.forEach(i => { if (i.name == bookmark.category) i.count-- })
      commit('categories', categories)
      commit('bookmarks', state.bookmarks.filter(i => i.id != bookmark.id))
    }
  }
})
app.use(store)

const routes = [
  { path: '/', component: showBookmarks },
  { path: '/setting', component: setting },
  { path: '/category/:mode', component: category },
  { path: '/bookmark/:mode', component: bookmark }
]

const router = VueRouter.createRouter({
  history: VueRouter.createWebHistory(),
  routes
})
app.use(router)

app.mixin({
  methods: {
    goback: function (reload) {
      if (reload)
        this.$store.dispatch('categories')
      this.$router.go(-1)
    }
  }
})

app.component('login', login)
app.component('sidebar', sidebar)

app.mount('#app')
