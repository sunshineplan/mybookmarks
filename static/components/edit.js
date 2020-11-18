const category = {
  data() {
    return {
      category: this.$store.state.editCategory,
      name: '',
      validated: false
    }
  },
  computed: {
    mode() {
      if (this.category.id == undefined) return 'Add'
      return 'Edit'
    }
  },
  template: `
<div @keyup.enter='save'>
  <header style='padding-left: 20px'>
    <h3>{{ mode }} Category</h3>
    <hr>
  </header>
  <div class='form' :class="{ 'was-validated': validated }">
    <div class='form-group'>
      <label for='category'>Category</label>
      <input class='form-control' v-model.trim='name' id='category' maxlength=15 required>
      <div class='invalid-feedback'>This field is required.</div>
      <small class='form-text text-muted'>Max length: 15 characters. One chinese character equal three characters.</small>
    </div>
    <button class='btn btn-primary' @click='save'>{{ mode }}</button>
    <button class='btn btn-primary' @click='goback'>Cancel</button>
  </div>
  <div class='form' v-if='category.id != undefined'>
    <button class='btn btn-danger delete' @click='del'>Delete</button>
  </div>
</div>`,
  created() { this.name = this.category.name },
  mounted() { document.title = this.mode + ' Category - My Bookmarks' },
  methods: {
    save: function () {
      if (valid()) {
        this.validated = false
        var r
        if (this.category.id == undefined)
          r = post('/category/add', { category: this.name })
        else
          r = post('/category/edit/' + this.category.id, { category: this.name })
        r.then(resp => {
          if (!resp.ok) resp.text().then(err =>
            BootstrapButtons.fire('Error', err, 'error'))
          else resp.json().then(json => {
            if (json.status == 1) {
              if (this.category.id != undefined)
                this.$store.commit('category', { id: this.category.id, name: this.name })
              this.$store.commit('categories')
              this.$store.commit('goto', 'showBookmark')
            }
            else BootstrapButtons.fire('Error', json.message, 'error')
          })
        })
      }
      else this.validated = true
    },
    del: function () {
      confirm('category').then(confirm => {
        if (confirm) post('/category/delete/' + this.category.id)
          .then(resp => {
            if (!resp.ok) resp.text().then(err =>
              BootstrapButtons.fire('Error', err, 'error'))
            else {
              this.$store.commit('category', { id: -1, name: 'All Bookmarks' })
              this.$store.commit('categories')
              this.$store.commit('goto', 'showBookmark')
            }
          })
      })
    }
  }
}

const bookmark = {
  data() {
    return {
      categories: this.$store.state.categories,
      bookmark: this.$store.state.bookmark,
      name: '',
      url: '',
      category: '',
      validated: false
    }
  },
  computed: {
    mode() {
      if (this.bookmark.id == undefined) return 'Add'
      return 'Edit'
    }
  },
  template: `
  <div @keyup.enter='save'>
    <header style='padding-left: 20px'>
      <h3>{{ mode }} Bookmark</h3>
      <hr>
    </header>
    <div class='form' :class="{ 'was-validated': validated }">
      <div class='form-group'>
        <label for='bookmark'>Bookmark</label>
        <input class='form-control' v-model.trim='name' id='bookmark' maxlength=40 required>
        <div class='invalid-feedback'>This field is required.</div>
        <small class='form-text text-muted'>Max length: 40 characters.</small>
      </div>
      <div class='form-group'>
        <label for='url'>URL</label>
        <input class='form-control' type='url' v-model.trim='url' id='url' @blur='chkURL' required>
        <div class='invalid-feedback'>Please enter a valid URL.</div>
      </div>
      <div class='form-group'>
        <label for='category'>Category</label>
        <input class='form-control' list='category-list' v-model.trim='category' id='category' maxlength=15>
        <datalist id='category-list'>
          <option v-for='c in categories'>{{ c.name }}</option>
        </datalist>
        <small class='form-text text-muted'>Max length: 15 characters. One chinese character equal three characters.</small>
      </div>
      <button class='btn btn-primary' @click='save'>{{ mode }}</button>
      <button class='btn btn-primary' @click='goback'>Cancel</button>
    </div>
    <div class='form' v-if='bookmark.id != undefined'>
      <button class='btn btn-danger delete' @click='del'>Delete</button>
    </div>
  </div>`,
  created() {
    this.name = this.bookmark.name
    this.url = this.bookmark.url
    this.category = this.bookmark.category
  },
  mounted() { document.title = this.mode + ' Bookmark - My Bookmarks' },
  methods: {
    chkURL: function () {
      if (this.url && !this.url.match(/^https?:/) && this.url.length)
        this.url = 'http://' + this.url
    },
    save: function () {
      if (valid()) {
        this.validated = false
        var r
        if (this.bookmark.id == undefined)
          r = post('/bookmark/add', {
            bookmark: this.name,
            url: this.url
          })
        else
          r = post('/bookmark/edit/' + this.bookmark.id, {
            bookmark: this.name,
            url: this.url
          })
        r.then(resp => {
          if (!resp.ok) resp.text().then(err =>
            BootstrapButtons.fire('Error', err, 'error'))
          else resp.json().then(json => {
            if (json.status == 1) this.$store.commit('goto', 'showBookmark')
            else BootstrapButtons.fire('Error', json.message, 'error')
              .then(() => {
                if (json.error == 1) this.name = ''
                else if (json.error == 2) this.url = ''
              })
          })
        })
      }
      else this.validated = true
    },
    del: function () {
      confirm('bookmark').then(confirm => {
        if (confirm) post('/bookmark/delete/' + this.bookmark.id)
          .then(resp => {
            if (!resp.ok) resp.text().then(err =>
              BootstrapButtons.fire('Error', err, 'error'))
            else this.$store.commit('goto', 'showBookmark')
          })
      })
    }
  }
}
