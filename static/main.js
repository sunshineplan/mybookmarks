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
      category: { id: -1, name: 'All Bookmarks' },
      bookmark: {},
      editCategory: {}
    }
  },
  mutations: {
    goto(state, component) { state.component = component },
    loading(state, status) { state.loading = status },
    ready(state) { state.sidebar = true },
    categories(stat) {
      stat.sidebar = false
      post('/category/get')
        .then(response => response.json())
        .then(json => {
          stat.categories = json
          stat.sidebar = true
          stat.loading = false
        })
    },
    category(state, category) { state.category = category },
    bookmark(state, bookmark) { state.bookmark = bookmark },
    editCategory(state, category) { state.editCategory = category }
  }
})
app.use(store)

app.mixin({ methods: { goback: function () { this.$store.commit('goto', 'showBookmark') } } })

app.component('login', login)
app.component('setting', setting)
app.component('sidebar', sidebar)
app.component('showBookmark', showBookmarks)
app.component('category', category)
app.component('bookmark', bookmark)

app.mount('#app')
