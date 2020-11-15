const bookmark = Vue.createApp({
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

bookmark.component('login', {
  data() {
    return {
      username: '',
      password: '',
      rememberme: false,
      validated: false
    }
  },
  template: `
<div @keyup.enter='login'>
  <header>
    <h3 class='d-flex justify-content-center align-items-center' style='height: 100%;'>Log In</h3>
  </header>
  <div class='login' :class="{ 'was-validated': validated }">
    <div class='form-group'>
      <label for='username'>Username</label>
      <input autofocus class='form-control' v-model='username' id='username' maxlength=20 placeholder='Username' required>
    </div>
    <div class='form-group'>
      <label for='password'>Password</label>
      <input class='form-control' type='password' v-model='password' id='password' maxlength=20 placeholder='Password' required>
    </div>
    <div class='form-group form-check'>
      <input type='checkbox' class='form-check-input' v-model='rememberme' id='rememberme'>
      <label class='form-check-label' for='rememberme'>Remember Me</label>
    </div>
    <hr>
    <button class='btn btn-primary login' @click='login'>Log In</button>
  </div>
</div>`,
  mounted() { document.title = 'Log In' },
  methods: {
    login: function () {
      if (valid()) {
        this.validated = false
        post('/login', {
          username: this.username,
          password: this.password,
          rememberme: this.rememberme
        }).then(resp => {
          if (!resp.ok) resp.text().then(err =>
            BootstrapButtons.fire('Error', err, 'error'))
          else window.location = '/'
        }).catch(e => BootstrapButtons.fire('Error', e, 'error'))
      }
      else this.validated = true
    }
  }
})

bookmark.component('setting', {
  data() {
    return {
      password: '',
      password1: '',
      password2: '',
      validated: false
    }
  },
  template: `
<div>
  <header style='padding-left: 20px;'>
    <h3>Setting</h3>
    <hr>
  </header>
  <div class='form' :class="{ 'was-validated': validated }">
    <div class='form-group'>
      <label for='password'>Current Password</label>
      <input class='form-control' type='password' v-model='password' id='password' maxlength=20 required>
      <div class='invalid-feedback'>This field is required.</div>
    </div>
    <div class='form-group'>
      <label for='password1'>New Password</label>
      <input class='form-control' type='password' v-model='password1' id='password1' maxlength=20 required>
      <div class='invalid-feedback'>This field is required.</div>
    </div>
    <div class='form-group'>
      <label for='password2'>Confirm Password</label>
      <input class='form-control' type='password' v-model='password2' id='password2' maxlength=20 required>
      <div class='invalid-feedback'>This field is required.</div>
      <small class='form-text text-muted'>Max password length: 20 characters.</small>
    </div>
    <button class='btn btn-primary' @click='setting'>Change</button>
    <button class='btn btn-primary' @click='goback'>Cancel</button>
  </div>
</div>`,
  mounted() { document.title = 'Setting' },
  methods: {
    setting: function () {
      if (valid()) {
        this.validated = false
        post('/setting', {
          password: this.password,
          password1: this.password1,
          password2: this.password2
        }).then(resp => {
          if (!resp.ok) resp.text().then(err =>
            BootstrapButtons.fire('Error', err, 'error'))
          else resp.json().then(json => {
            if (json.status == 1)
              BootstrapButtons.fire('Success', 'Your password has changed. Please Re-login!', 'success')
                .then(() => window.location = '/')
            else
              BootstrapButtons.fire('Error', json.message, 'error')
                .then(() => {
                  if (json.error == 1) this.password = ''
                  else {
                    this.password1 = ''
                    this.password2 = ''
                  }
                })
          })
        }).catch(e => BootstrapButtons.fire('Error', e, 'error'))
      }
      else this.validated = true
    },
    goback: function () { this.$parent.content = 'showBookmark' }
  }
})

bookmark.component('sidebar', {
  props: { active: Number },
  data() { return { category: {} } },
  template: `
<nav class='nav flex-column navbar-light sidebar'>
  <div class='category-menu'>
    <button class='btn btn-primary btn-sm' @click='add'>Add Category</button>
    <ul class='navbar-nav' v-if='category.total'>
      <a
        class='navbar-brand category'
        :class='{ active: active == -1 }'
        @click="load(-1, 'All Bookmarks')"
      >
        All Bookmarks ({{ category.total }})
      </a>
      <li v-for='c in category.categories'>
        <a
          class='nav-link category'
          :class='{ active: active == c.ID }'
          @click='load(c.ID, c.Name)'
        >
          {{ c.Name }} ({{ c.Count }})
        </a>
      </li>
      <li>
        <a
          class='nav-link category'
          :class='{ active: active == 0 }'
          @click="load(0, 'Uncategorized')"
        >
          Uncategorized ({{ category.uncategorized }})
        </a>
      </li>
    </ul>
  </div>
</nav>`,
  created() {
    this.$parent.loading = true
    post('/category/get')
      .then(response => response.json())
      .then(json => {
        this.category = json
        this.$parent.siderbar = true
        this.$parent.loading = false
      })
  },
  mounted() {
    window.addEventListener('keyup', this.arrow)
  },
  beforeUnmount: function () {
    window.removeEventListener('keyup', this.arrow)
  },
  methods: {
    arrow: function (event) {
      if (this.active != null) {
        var len = this.category.categories.length
        var index = this.category.categories.findIndex(item => item.ID == this.active)
        if (event.key == 'ArrowUp') {
          if (this.active == 0 && len > 0)
            this.load(this.category.categories[len - 1].ID, this.category.categories[len - 1].Name)
          else if (index > 0)
            this.load(this.category.categories[index - 1].ID, this.category.categories[index - 1].Name)
          else if (index == 0) this.load(-1, 'All Bookmarks')
        } else if (event.key == 'ArrowDown')
          if (this.active == -1 && len > 0)
            this.load(this.category.categories[0].ID, this.category.categories[0].Name)
          else if (index >= 0 && index < len - 1)
            this.load(this.category.categories[index + 1].ID, this.category.categories[index + 1].Name)
          else if (index == len - 1) this.load(0, 'Uncategorized')
      }

    },
    add: function () {
      this.$parent.category = {}
      this.$parent.content = 'category'
    },
    load: function (id, category) {
      this.$parent.content = 'showBookmark'
      this.$parent.current = { id: id, category: category }
    }
  }
})

bookmark.component('showBookmark', {
  props: {
    current: Object
  },
  data() {
    return {
      bookmark: {
        bookmarks: [],
        category: { name: this.current.category }
      },
      smallSize: window.innerWidth <= 700 ? true : false
    }
  },
  template: `
<div>
  <header style='padding-left: 20px;'>
    <div style='height: 50px;'>
      <a class='h3 title'>{{ bookmark.category.name }}</a>
      <a class='btn icon' v-if='bookmark.category.id > 0' @click='editCategory'>
        <i class='material-icons edit'>edit</i>
      </a>
    </div>
    <button class='btn btn-primary' @click='add'>Add Bookmark</button>
  </header>
  <div class='table-responsive'>
    <table class='table table-sm'>
      <thead>
        <tr>
          <th>Bookmark</th>
          <th>URL</th>
          <th>Category</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        <tr v-for='b in bookmark.bookmarks'>
          <td>{{ b.Name }}</td>
          <td><a :href='b.URL' target='_blank' class='url' :data-url='b.URL'>{{ b.URL }}</a></td>
          <td>{{ b.Category }}</td>
          <td>
            <a class='icon' @click='edit(b)'><i class='material-icons edit'>edit</i></a>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</div>`,
  mounted() {
    this.load(this.$parent.current.id)
    window.addEventListener('resize', this.checkSize)
  },
  beforeUnmount: function () {
    window.removeEventListener('resize', this.checkSize)
  },
  watch: {
    current(obj) { this.load(obj.id) },
    smallSize(isSmall) {
      if (isSmall) Array.from(document.getElementsByClassName('url'))
        .forEach(i => i.text = i.text.replace(/https?:\/\/(www\.)?/i, ''))
      else Array.from(document.getElementsByClassName('url'))
        .forEach(i => i.text = i.dataset.url)
    }
  },
  methods: {
    load: function (id) {
      this.$parent.active = this.$parent.current.id
      this.$parent.loading = true
      post('/bookmark/get', { category: id })
        .then(resp => {
          if (!resp.ok) resp.text().then(err =>
            BootstrapButtons.fire('Error', err, 'error'))
          else resp.json().then(json => {
            this.bookmark = json
            this.$parent.loading = false
            document.title = this.current.category + ' - My Bookmarks'
          })
        })
        .catch(e => BootstrapButtons.fire('Error', e, 'error'))
        .then(() => this.$parent.loading = false)
    },
    checkSize: function () {
      if (window.innerWidth <= 700) this.smallSize = true
      else this.smallSize = false
    },
    editCategory: function () {
      this.$parent.category = {
        Name: this.bookmark.category.name,
        ID: this.bookmark.category.id
      }
      this.$parent.content = 'category'
    },
    add: function () {
      this.$parent.bookmark = {}
      this.$parent.content = 'bookmark'
    },
    edit: function (bookmark) {
      this.$parent.bookmark = bookmark
      this.$parent.content = 'bookmark'
    }
  }
})

bookmark.component('category', {
  props: { category: Object },
  data() {
    return {
      name: this.category.Name,
      validated: false
    }
  },
  template: `
<div>
  <header style='padding-left: 20px;'>
    <a class='h3 title'>{{ mode }} Category</a>
    <hr>
  </header>
  <div class='form' :class="{ 'was-validated': validated }">
    <div class='form-group'>
      <label for='category'>Category</label>
      <input class='form-control' v-model='name' id='category' maxlength=15 required>
      <div class='invalid-feedback'>This field is required.</div>
      <small class='form-text text-muted'>Max length: 15 characters. One chinese character equal three characters.</small>
    </div>
    <button class='btn btn-primary' @click='save'>{{ mode }}</button>
    <button class='btn btn-primary' @click='goback'>Cancel</button>
  </div>
  <div class='form' v-if='category.ID != 0'>
    <button class='btn btn-danger delete' @click='del'>Delete</button>
  </div>
</div>`,
  mounted() { document.title = this.mode + ' Category - My Bookmarks' },
  methods: {
    save: function () {
      if (valid()) {
        this.validated = false
        var r
        if (this.category.ID == undefined)
          r = post('/category/add', { category: this.name })
        else
          r = post('/category/edit/' + this.category.ID, { category: this.name })
        r.then(resp => {
          if (!resp.ok) resp.text().then(err =>
            BootstrapButtons.fire('Error', err, 'error'))
          else resp.json().then(json => {
            if (json.status == 1) this.$parent.content = 'showBookmark'
            else BootstrapButtons.fire('Error', json.message, 'error')
          })
        }).catch(e => BootstrapButtons.fire('Error', e, 'error'))
      }
      else this.validated = true
    },
    del: function () {
      confirm('category').then(confirm => {
        if (confirm) post('/category/delete/' + this.category.ID)
          .then(resp => {
            if (!resp.ok) resp.text().then(err =>
              BootstrapButtons.fire('Error', err, 'error'))
            else this.$parent.content = 'showBookmark'
          })
          .catch(e => BootstrapButtons.fire('Error', e, 'error'))
      })
    },
    goback: function () { this.$parent.content = 'showBookmark' }
  },
  computed: {
    mode: function () {
      if (this.category.ID == undefined) return 'Add'
      return 'Edit'
    }
  }
})

bookmark.component('bookmark', {
  props: {
    bookmark: Object,
    categories: Array
  },
  data() {
    return {
      name: this.bookmark.Name,
      url: this.bookmark.URL,
      category: this.bookmark.Category,
      validated: false
    }
  },
  template: `
<div>
  <header style='padding-left: 20px;'>
    <a class='h3 title'>{{ mode }} Bookmark</a>
    <hr>
  </header>
  <div class='form' :class="{ 'was-validated': validated }">
    <div class='form-group'>
      <label for='bookmark'>Bookmark</label>
      <input class='form-control' v-model='name' id='bookmark' maxlength=40 required>
      <div class='invalid-feedback'>This field is required.</div>
      <small class='form-text text-muted'>Max length: 40 characters.</small>
    </div>
    <div class='form-group'>
      <label for='url'>URL</label>
      <input class='form-control' type='url' v-model='url' id='url' @blur='chkURL' required>
      <div class='invalid-feedback'>Please enter a valid URL.</div>
    </div>
    <div class='form-group'>
      <label for='category'>Category</label>
      <input class='form-control' list='category-list' v-model='category' id='category' maxlength=15>
      <datalist id='category-list'>
        <option v-for='c in categories'>{{ c.Name }}</option>
      </datalist>
      <small class='form-text text-muted'>Max length: 15 characters. One chinese character equal three characters.</small>
    </div>
    <button class='btn btn-primary' @click='save'>{{ mode }}</button>
    <button class='btn btn-primary' @click='goback'>Cancel</button>
  </div>
  <div class='form' v-if='bookmark.ID != 0'>
    <button class='btn btn-danger delete' @click='del'>Delete</button>
  </div>
</div>`,
  mounted() { document.title = this.mode + ' Bookmark - My Bookmarks' },
  methods: {
    chkURL: function () {
      if (!this.url.match(/^https?:/) && this.url.length)
        this.url = 'http://' + this.url
    },
    save: function () {
      if (valid()) {
        this.validated = false
        var r
        if (this.bookmark.ID == undefined)
          r = post('/bookmark/add', {
            bookmark: this.name,
            url: this.url
          })
        else
          r = post('/bookmark/edit/' + this.bookmark.ID, {
            bookmark: this.name,
            url: this.url
          })
        r.then(resp => {
          if (!resp.ok) resp.text().then(err =>
            BootstrapButtons.fire('Error', err, 'error'))
          else resp.json().then(json => {
            if (json.status == 1) this.$parent.content = 'showBookmark'
            else BootstrapButtons.fire('Error', json.message, 'error')
              .then(() => {
                if (json.error == 1) this.name = ''
                else if (json.error == 2) this.url = ''
              })
          })
        }).catch(e => BootstrapButtons.fire('Error', e, 'error'))
      }
      else this.validated = true
    },
    del: function () {
      confirm('category').then(confirm => {
        if (confirm) post('/bookmark/delete/' + this.bookmark.ID)
          .then(resp => {
            if (!resp.ok) resp.text().then(err =>
              BootstrapButtons.fire('Error', err, 'error'))
            else this.$parent.content = 'showBookmark'
          })
          .catch(e => BootstrapButtons.fire('Error', e, 'error'))
      })
    },
    goback: function () { this.$parent.content = 'showBookmark' }
  },
  computed: {
    mode: function () {
      if (this.bookmark.ID == undefined) return 'Add'
      return 'Edit'
    }
  }
})

bookmark.mount('#bookmark')
