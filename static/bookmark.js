post = (url, obj) => {
  return fetch(url, {
    method: 'post',
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded'
    },
    body: new URLSearchParams(obj)
  })
}

const bookmark = Vue.createApp({
  data() {
    return {
      user: document.getElementById('user').value,
      content: 'showBookmark',
      current: -1,
      loading: false,
      category: {},
      bookmark: {}
    }
  },
  computed: {
    prop: function () {
      if (this.content == 'showBookmark')
        return { current: 'current' }
      else if (this.content == 'category')
        return { category: 'category' }
      else if (this.content == 'bookmark')
        return { bookmark: 'bookmark' }
    }
  }
})

bookmark.component('login', {
  data() {
    return {
      username: '',
      password: '',
      rememberme: false
    }
  },
  template: `
<div>
  <header>
    <h3 class='d-flex justify-content-center align-items-center' style='height: 100%;'>Log In</h3>
  </header>
  <div class='login'>
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
      post('/login', {
        username: this.username,
        password: this.password,
        rememberme: this.rememberme
      }).then(resp => {
        if (!resp.ok)
          resp.text().then(err =>
            BootstrapButtons.fire('Error', err, 'error'))
        else
          window.location = '/'
      }).catch(e =>
        BootstrapButtons.fire('Error', e, 'error'))
    }
  }
})

bookmark.component('setting', {
  data() {
    return {
      password: '',
      password1: '',
      password2: ''
    }
  },
  template: `
<div>
  <header style='padding-left: 20px;'>
    <h3>Setting</h3>
    <hr>
  </header>
  <div class='form'>
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
      post('/setting', {
        password: this.password,
        password1: this.password1,
        password2: this.password2
      }).then(resp => {
        if (!resp.ok)
          resp.text().then(err =>
            BootstrapButtons.fire('Error', err, 'error'))
        else
          resp.json().then(json => {
            if (json.status == 1)
              BootstrapButtons.fire('Success', 'Your password has changed. Please Re-login!', 'success')
                .then(() => window.location = '/')
            else
              BootstrapButtons.fire('Error', json.message, 'error')
                .then(() => {
                  if (json.error == 1)
                    this.password = ''
                  else {
                    this.password1 = ''
                    this.password2 = ''
                  }
                })
          })
      }).catch(e =>
        BootstrapButtons.fire('Error', e, 'error'))
    },
    goback: function () { this.$parent.content = 'showBookmark' }
  }
})

bookmark.component('sidebar', {
  data() {
    return {
      category: {}
    }
  },
  template: `
<nav class='nav flex-column navbar-light sidebar'>
  <div class='category-menu'>
    <button class='btn btn-primary btn-sm' @click='add'>Add Category</button>
    <a class='navbar-brand category' @click='load(-1)'>All Bookmarks ({{ category.total }})</a>
    <ul class='navbar-nav'>
      <li v-for='c in category.categories'>
        <a class='nav-link category' @click='load(c.ID)'>{{ c.Name }} ({{ c.Count }})</a>
      </li>
      <li>
        <a class='nav-link category' @click='load(0)'>Uncategorized ({{ category.uncategorized }})</a>
      </li>
    </ul>
  </div>
</nav>`,
  created() {
    post('/category/get')
      .then(response => response.json())
      .then(json => { this.category = json })
  },
  methods: {
    add: function () { this.$parent.content = 'category' },
    load: function (id) { this.$parent.current = id }
  }
})

bookmark.component('showBookmark', {
  props: {
    current: Number
  },
  data() {
    return {
      bookmark: {}
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
          <th>{{ b.Name }}</th>
          <th><a :href='b.URL' target='_blank' class='url'></a></th>
          <th>{{ b.Category }}</th>
          <th>
            <a class='icon' @click='edit(b)'><i class='material-icons edit'>edit</i></a>
          </th>
        </tr>
      </tbody>
    </table>
  </div>
</div>`,
  created() {
    load(this.$parent.current)
  },
  watch: {
    current(id) { load(id) }
  },
  methods: {
    load: function (id) {
      post('/bookmark/get', { category: id })
        .then(response => response.json())
        .then(json => {
          this.bookmark = json
          document.title = this.bookmark.category.name + ' - My Bookmarks'
        })
    },
    editCategory: function () {
      this.$parent.category = {
        Name: this.bookmark.category.name,
        ID: this.bookmark.category.id
      }
      this.$parent.content = 'category'
    },
    add: function () { this.$parent.content = 'bookmark' },
    edit: function (bookmark) {
      this.$parent.bookmark = bookmark
      this.$parent.content = 'bookmark'
    }
  }
})

bookmark.component('category', {
  props: {
    category: Object
  },
  data() {
    return {
      name: this.category.Name
    }
  },
  template: `
<div>
  <header style='padding-left: 20px;'>
    <a class='h3 title'>{{ mode }} Category</a>
    <hr>
  </header>
  <div class='form'>
    <div class='form-group'>
      <label for='category'>Category</label>
      <input class='form-control' v-model='name' id='category' maxlength=15' required>
      <div class='invalid-feedback'>This field is required.</div>
      <small class='form-text text-muted'>Max length: 15 characters. One chinese character equal three characters.</small>
    </div>
    <button class='btn btn-primary' @click='do'>{{ mode }}</button>
    <button class='btn btn-primary' @click='goback'>Cancel</button>
  </div>
  <div class='form' v-if='category.ID != 0'>
    <button class='btn btn-danger delete' @click='delete'>Delete</button>
  </div>
</div>`,
  mounted() { document.title = this.mode + ' Category - My Bookmarks' },
  methods: {
    do: function () {
      var r
      if (this.category.ID == 0)
        r = post('/category/add', { category: this.name })
      else
        r = post('/category/edit' + this.category.ID, { category: this.name })
      r.then(resp => {
        if (!resp.ok)
          resp.text().then(err =>
            BootstrapButtons.fire('Error', err, 'error'))
        else
          resp.json().then(json => {
            if (json.status == 1)
              this.$parent.content = 'showBookmark'
            else
              BootstrapButtons.fire('Error', json.message, 'error')
          })
      }).catch(e =>
        BootstrapButtons.fire('Error', e, 'error'))
    },
    delete: function () {
      post('/category/delete' + this.category.ID)
        .then(resp => {
          if (!resp.ok)
            resp.text().then(err =>
              BootstrapButtons.fire('Error', err, 'error'))
          else
            this.$parent.content = 'showBookmark'
        }).catch(e =>
          BootstrapButtons.fire('Error', e, 'error'))
    },
    goback: function () { this.$parent.content = 'showBookmark' }
  },
  computed: {
    mode: function () {
      if (this.category.ID == 0)
        return 'Add'
      else return 'Edit'
    }
  }
})

bookmark.component('bookmark', {
  props: {
    bookmark: Object
  },
  data() {
    return {
      name: this.bookmark.Name,
      url: this.bookmark.URL,
      category: this.bookmark.Category
    }
  },
  template: `
<div>
  <header style='padding-left: 20px;'>
    <a class='h3 title'>{{ mode }} Bookmark</a>
    <hr>
  </header>
  <div class='form'>
    <div class='form-group'>
      <label for='bookmark'>Bookmark</label>
      <input class='form-control' v-model='bookmark' id='bookmark' maxlength=40 required>
      <div class='invalid-feedback'>This field is required.</div>
      <small class='form-text text-muted'>Max length: 40 characters.</small>
    </div>
    <div class='form-group'>
      <label for='url'>URL</label>
      <input class='form-control' type='url' v-model='url' id='url' required>
      <div class='invalid-feedback'>Please enter a valid URL.</div>
    </div>
    <div class='form-group'>
      <label for='category'>Category</label>
      <input class='form-control' list='category-list' v-model='category' id='category' maxlength=15>
      <datalist id='category-list'>
        <option v-for='c in $ref.categories.category.categories'>{{ c.Name }}</option>
      </datalist>
      <small class='form-text text-muted'>Max length: 15 characters. One chinese character equal three characters.</small>
    </div>
    <button class='btn btn-primary' @click='do'>{{ mode }}</button>
    <button class='btn btn-primary' @click='goback'>Cancel</button>
  </div>
  <div class='form' v-if='bookmark.ID != 0'>
    <button class='btn btn-danger delete' onclick='delete'>Delete</button>
  </div>
</div>`,
  mounted() { document.title = this.mode + ' Bookmark - My Bookmarks' },
  methods: {
    do: function () {
      var r
      if (this.bookmark.ID == 0)
        r = post('/bookmark/add', {
          bookmark: this.name,
          url: this.url
        })
      else
        r = post('/bookmark/edit' + this.bookmark.ID, {
          bookmark: this.name,
          url: this.url
        })
      r.then(resp => {
        if (!resp.ok)
          resp.text().then(err =>
            BootstrapButtons.fire('Error', err, 'error'))
        else
          resp.json().then(json => {
            if (json.status == 1)
              this.$parent.content = 'showBookmark'
            else
              BootstrapButtons.fire('Error', json.message, 'error')
          })
      }).catch(e =>
        BootstrapButtons.fire('Error', e, 'error'))
    },
    delete: function () {
      post('/bookmark/delete' + this.bookmark.ID)
        .then(resp => {
          if (!resp.ok)
            resp.text().then(err =>
              BootstrapButtons.fire('Error', err, 'error'))
          else
            this.$parent.content = 'showBookmark'
        }).catch(e =>
          BootstrapButtons.fire('Error', e, 'error'))
    },
    goback: function () { this.$parent.content = 'showBookmark' }
  },
  computed: {
    mode: function () {
      if (this.bookmark.ID == 0)
        return 'Add'
      else return 'Edit'
    }
  }
})

bookmark.mount('#bookmark')
