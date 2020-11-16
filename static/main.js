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
