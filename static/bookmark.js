const bookmark = Vue.createApp({
  data() {
    return {
      user: document.getElementById('user').value,
      content: 'showBookmark',
      last: -1,
      loading: false
    }
  },
  computed: {
    prop: function () {
      if (this.content == 'category')
        return { bookmark: 'category' }
      else if (this.content == 'bookmark')
        return { bookmark: 'bookmark' }
    }
  }
})

bookmark.component('login', {
  template: `
<div>
  <header>
    <h3 class='d-flex justify-content-center align-items-center' style='height: 100%;'>Log In</h3>
  </header>
  <div class='login'>
    <div class='form-group'>
      <label for='username'>Username</label>
      <input autofocus class='form-control' name='username' id='username' maxlength=20 placeholder='Username' required>
    </div>
    <div class='form-group'>
      <label for='password'>Password</label>
      <input class='form-control' type='password' name='password' id='password' maxlength=20 placeholder='Password' required>
    </div>
    <div class='form-group form-check'>
      <input type='checkbox' class='form-check-input' name='rememberme' id='rememberme'>
      <label class='form-check-label' for='rememberme'>Remember Me</label>
    </div>
    <hr>
    <button class='btn btn-primary login' @click='login' id='submit'>Log In</button>
  </div>
</div>`,
  mounted() { document.title = 'Log In' },
  methods: {
    login: () => { }
  }
})

bookmark.component('setting', {
  template: `
<div>
  <header style='padding-left: 20px;'>
    <h3>Setting</h3>
    <hr>
  </header>
  <div class='form'>
    <div class='form-group'>
      <label for='password'>Current Password</label>
      <input class='form-control' type='password' name='password' id='password' maxlength=20 required>
      <div class='invalid-feedback'>This field is required.</div>
    </div>
    <div class='form-group'>
      <label for='password1'>New Password</label>
      <input class='form-control' type='password' name='password1' id='password1' maxlength=20 required>
      <div class='invalid-feedback'>This field is required.</div>
    </div>
    <div class='form-group'>
      <label for='password2'>Confirm Password</label>
      <input class='form-control' type='password' name='password2' id='password2' maxlength=20 required>
      <div class='invalid-feedback'>This field is required.</div>
      <small class='form-text text-muted'>Max password length: 20 characters.</small>
    </div>
    <button class='btn btn-primary' @click='setting' id='submit'>Change</button>
    <button class='btn btn-primary' @click='goback'>Cancel</button>
  </div>
</div>`,
  mounted() { document.title = 'Setting' },
  methods: {
    setting: () => { },
    goback: () => { }
  }
})

bookmark.component('sidebar', {
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
  data() {
    return {
      category: {}
    }
  },
  created() {
    fetch('/category/get')
      .then(response => response.json())
      .then(json => { this.category = json })
  },
  methods: {
    add: () => { },
    load: id => { console.log(id) }
  }
})

bookmark.component('showBookmark', {
  template: `
<div>
  <header style='padding-left: 20px;'>
    <div style='height: 50px;'>
      <a class='h3 title'>{{ bookmark.category.name }}</a>
      <a class='btn icon' v-if='bookmark.category.id > 0' @click='editCategory(bookmark.category.id)'>
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
            <a class='icon' @click='edit(b.ID)'><i class='material-icons edit'>edit</i></a>
          </th>
        </tr>
      </tbody>
    </table>
  </div>
</div>`,
  mounted() { document.title = this.bookmark.category.name + ' - My Bookmarks' },
  data() {
    return {
      bookmark: {}
    }
  },
  created() {
    fetch('/bookmark/get')
      .then(response => response.json())
      .then(json => { this.bookmark = json })
  },
  methods: {
    editCategory: id => { console.log(id) },
    add: () => { },
    edit: id => { console.log(id) }
  }
})

bookmark.component('category', {
  props: {
    category: Object
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
      <input class='form-control' name='category' id='category' maxlength=15 :value='category.Name' required>
      <div class='invalid-feedback'>This field is required.</div>
      <small class='form-text text-muted'>Max length: 15 characters. One chinese character equal three characters.</small>
    </div>
    <button class='btn btn-primary' @click='do({{ category.ID }})' id='submit'>{{ mode }}</button>
    <button class='btn btn-primary' @click='goback'>Cancel</button>
  </div>
  <div class='form' v-if='category.ID != 0'>
    <button class='btn btn-danger delete' @click='delete({{ category.ID }})'>Delete</button>
  </div>
</div>`,
  mounted() { document.title = this.mode + ' Category - My Bookmarks' },
  methods: {
    do: id => { console.log(id) },
    delete: id => { console.log(id) },
    goback: () => { }
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
  template: `
<div>
  <header style='padding-left: 20px;'>
    <a class='h3 title'>{{ mode }} Bookmark</a>
    <hr>
  </header>
  <div class='form'>
    <div class='form-group'>
      <label for='bookmark'>Bookmark</label>
      <input class='form-control' name='bookmark' id='bookmark' maxlength=40 :value='bookmark.Name' required>
      <div class='invalid-feedback'>This field is required.</div>
      <small class='form-text text-muted'>Max length: 40 characters.</small>
    </div>
    <div class='form-group'>
      <label for='url'>URL</label>
      <input class='form-control' type='url' name='url' id='url' :value='bookmark.URL' required>
      <div class='invalid-feedback'>Please enter a valid URL.</div>
    </div>
    <div class='form-group'>
      <label for='category'>Category</label>
      <input class='form-control' list='category-list' name='category' id='category' maxlength=15 :value='bookmark.Category'>
      <datalist id='category-list'>
      {{ range $_, $category := .categories }}
        <option value="{{ $category }}"></option>
      {{ end }}
      </datalist>
      <small class='form-text text-muted'>Max length: 15 characters. One chinese character equal three characters.</small>
    </div>
    <button class='btn btn-primary' @click='do({{ bookmark.ID }})' id='submit'>{{ mode }}</button>
    <button class='btn btn-primary' @click='goback'>Cancel</button>
  </div>
  <div class='form' v-if='bookmark.ID != 0'>
    <button class='btn btn-danger delete' onclick='delete({{ bookmark.ID }})'>Delete</button>
  </div>
</div>`,
  mounted() { document.title = this.mode + ' Bookmark - My Bookmarks' },
  methods: {
    do: id => { console.log(id) },
    delete: id => { console.log(id) },
    goback: () => { }
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